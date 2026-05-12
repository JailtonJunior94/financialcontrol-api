package migration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/database/migration"
	localdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
)

// smokeTriggerVersion is the migration version after which the smoke hook is invoked.
// Version 2 corresponds to 000002_finance_module_ddl_dml (DDL + DML + gate).
const smokeTriggerVersion = 2

// SmokeHookFn is a function invoked between the smoke-trigger migration and the
// subsequent DROP migration.  Implementations must write Status='smoke_ok' to
// finance.MigrationAudit before returning nil.
type SmokeHookFn func(ctx context.Context, logger *slog.Logger) error

// globalHooks is the package-level registry populated by hook packages via init().
var globalHooks = map[string]SmokeHookFn{}

// RegisterHook registers a named smoke hook in the global registry.
// Called by hook packages in their init() functions.
func RegisterHook(name string, fn SmokeHookFn) {
	globalHooks[name] = fn
}

// Runner is the entrypoint called by cmd/migration/main.go.
type Runner interface {
	Run(ctx context.Context) error
}

// Option configures the runner (used in tests for dependency injection).
type Option func(*runner)

type runner struct {
	cfg         Config
	logger      *slog.Logger
	open        func(context.Context, string, *slog.Logger) (manager.Manager, error)
	factory     func(manager.Manager) (migration.Migrator, error)
	smokeHook   string
	testHookFn  SmokeHookFn // injected in tests; overrides globalHooks lookup
}

// WithOpenFunc overrides the Manager open function (for tests).
func WithOpenFunc(fn func(context.Context, string, *slog.Logger) (manager.Manager, error)) Option {
	return func(r *runner) { r.open = fn }
}

// WithFactory overrides the Migrator factory (for tests).
func WithFactory(fn func(manager.Manager) (migration.Migrator, error)) Option {
	return func(r *runner) { r.factory = fn }
}

// WithLogger overrides the logger (for tests).
func WithLogger(log *slog.Logger) Option {
	return func(r *runner) { r.logger = log }
}

// WithSmokeHook sets the name of the smoke hook to invoke after the trigger migration.
// The hook must be registered via RegisterHook (typically from an init() in a hooks package).
func WithSmokeHook(name string) Option {
	return func(r *runner) { r.smokeHook = name }
}

// withTestHookFn injects a smoke hook function directly, bypassing the global registry.
// Intended for use in tests within this package.
func withTestHookFn(name string, fn SmokeHookFn) Option {
	return func(r *runner) {
		r.smokeHook = name
		r.testHookFn = fn
	}
}

// Run is the package-level entry point used by cmd/migration/main.go.
// It wires and runs the migration runner in one call.
func Run(ctx context.Context, opts ...Option) error {
	r, err := New(opts...)
	if err != nil {
		return err
	}
	return r.Run(ctx)
}

// New wires the runner with defaults: JSON logger, env-based config, retryable open.
// Returns an error if LoadConfig fails.
func New(opts ...Option) (Runner, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	r := &runner{
		cfg:    cfg,
		logger: NewJSONLogger(),
		open:   RetryOpen,
		factory: func(mgr manager.Manager) (migration.Migrator, error) {
			fs := localdatabase.MigrationsFS()
			return migration.New(mgr,
				migration.EmbedFS{FS: fs, Root: "migrations"},
				migration.WithDSN(cfg.DSN),
			)
		},
	}
	for _, o := range opts {
		o(r)
	}
	return r, nil
}

func (r *runner) Run(ctx context.Context) error {
	r.logStartup()

	ctx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	mgr, err := r.open(ctx, r.cfg.DSN, r.logger)
	if err != nil {
		return fmt.Errorf("open manager: %w", err)
	}
	defer shutdown(mgr, r.logger)

	if err := preflight(ctx, mgr); err != nil {
		return fmt.Errorf("preflight: %w", err)
	}

	migrator, err := r.factory(mgr)
	if err != nil {
		return fmt.Errorf("build migrator: %w", err)
	}

	return r.apply(ctx, migrator)
}

func (r *runner) logStartup() {
	attrs := append(BuildInfoFields(), AuditFields()...)
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	r.logger.Info("migration starting", args...)
}

func (r *runner) apply(ctx context.Context, migrator migration.Migrator) error {
	if r.cfg.BaselineVersion > 0 {
		return r.forceBaseline(ctx, migrator)
	}
	if r.smokeHook != "" {
		return r.applyWithSmoke(ctx, migrator)
	}
	return runUp(ctx, migrator, r.logger)
}

// applyWithSmoke implements the three-phase smoke pipeline:
//  1. Apply all pending migrations (000002 succeeds; 000003 SQL gate fires
//     because smoke_ok is absent, leaving the migration dirty at version 3).
//  2. Reset dirty state to smokeTriggerVersion, then run the named smoke hook
//     (which writes Status='smoke_ok' to finance.MigrationAudit).
//  3. Re-apply pending migrations (only 000003 remains; guard is now satisfied).
func (r *runner) applyWithSmoke(ctx context.Context, migrator migration.Migrator) error {
	hookFn := r.testHookFn
	if hookFn == nil {
		var ok bool
		hookFn, ok = globalHooks[r.smokeHook]
		if !ok {
			return fmt.Errorf("smoke hook %q not registered", r.smokeHook)
		}
	}

	// Phase 1: apply migrations; 000003 gate will fail if smoke hasn't run yet.
	// The error is expected — we log at Info level and continue.
	if firstErr := migrator.Up(ctx); firstErr != nil &&
		!errors.Is(firstErr, migration.ErrNoChange) {
		r.logger.Info("pre-smoke up stopped (gate may have fired)",
			slog.String("detail", firstErr.Error()))
	}

	ver, dirty, vErr := migrator.Version(ctx)
	if vErr != nil {
		return fmt.Errorf("version check after first up: %w", vErr)
	}

	// All migrations already applied in a previous run — nothing to do.
	if !dirty && ver > smokeTriggerVersion {
		r.logger.Info("migrations already complete, skipping smoke hook",
			slog.Uint64("version", uint64(ver)))
		return nil
	}

	// Fail fast if 000002 itself didn't complete.
	if ver < smokeTriggerVersion || (dirty && ver <= smokeTriggerVersion) {
		return fmt.Errorf(
			"migration stalled at version %d (dirty=%v) — expected >= %d",
			ver, dirty, smokeTriggerVersion)
	}

	// Reset dirty state left by the 000003 gate failure.
	if dirty {
		if err := migrator.Force(ctx, smokeTriggerVersion); err != nil {
			return fmt.Errorf("reset dirty migration to v%d: %w", smokeTriggerVersion, err)
		}
		r.logger.Info("reset dirty state",
			slog.Uint64("version", uint64(smokeTriggerVersion)))
	}

	// Phase 2: run smoke hook (writes smoke_ok to finance.MigrationAudit).
	r.logger.Info("invoking smoke hook", slog.String("hook", r.smokeHook))
	if err := hookFn(ctx, r.logger); err != nil {
		return fmt.Errorf("smoke %q: %w", r.smokeHook, err)
	}
	r.logger.Info("smoke hook passed", slog.String("hook", r.smokeHook))

	// Phase 3: apply 000003 (smoke_ok guard is now satisfied).
	return runUp(ctx, migrator, r.logger)
}

func (r *runner) forceBaseline(ctx context.Context, migrator migration.Migrator) error {
	if err := migrator.Force(ctx, int(r.cfg.BaselineVersion)); err != nil {
		return fmt.Errorf("baseline force: %w", err)
	}
	r.logger.Info("migration baseline forced", slog.Uint64("version", uint64(r.cfg.BaselineVersion)))
	return nil
}

func runUp(ctx context.Context, migrator migration.Migrator, log *slog.Logger) error {
	err := migrator.Up(ctx)
	if err == nil {
		log.Info("migration up complete")
		return nil
	}
	if errors.Is(err, migration.ErrNoChange) {
		log.Info("migration up no change")
		return nil
	}
	if errors.Is(err, ErrMigrationsDirEmpty) {
		log.Info("migration up no sql files found")
		return nil
	}
	log.Error("migration up failed", slog.String("error", err.Error()))
	return fmt.Errorf("migrate up: %w", err)
}

func preflight(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

func shutdown(mgr manager.Manager, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := mgr.Shutdown(ctx); err != nil {
		log.Warn("migration shutdown error", slog.String("error", err.Error()))
		return
	}
	log.Info("migration shutdown")
}
