//go:build integration

package migration_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"sync"
	"testing"
	"time"

	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	devkitmigration "github.com/JailtonJunior94/devkit-go/pkg/database/migration"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	migratorpkg "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration"
	migratorpkg_mocks "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/mocks"
	localdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

// RunnerIntegrationSuite covers the 8 CLI integration scenarios from techspec §Testes do CLI.
// D-57: one MSSQL testcontainer per suite; state reset between tests via SetupTest.
type RunnerIntegrationSuite struct {
	suite.Suite

	db  *sql.DB
	dsn string
}

func TestRunnerIntegration(t *testing.T) {
	suite.Run(t, new(RunnerIntegrationSuite))
}

func (s *RunnerIntegrationSuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to start shared MSSQL testcontainer")
	s.db = db

	dsn, err := dbmssql.GetSharedTestDSN()
	s.Require().NoError(err, "failed to get shared MSSQL DSN")
	s.dsn = dsn
}

// SetupTest drops migration-related tables to restore a clean DB state for each scenario.
func (s *RunnerIntegrationSuite) SetupTest() {
	ctx := context.Background()
	for _, stmt := range cleanupStatements {
		_, err := s.db.ExecContext(ctx, stmt)
		s.Require().NoError(err, "cleanup DDL: %s", stmt)
	}
}

// ---- Cenário 1: happy path em banco limpo (RNF-01) ----

func (s *RunnerIntegrationSuite) TestScenario1_HappyPath() {
	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "5m")
	s.T().Setenv("MIGRATION_BASELINE", "")

	r, err := migratorpkg.New(migratorpkg.WithLogger(discardLogger()))
	s.Require().NoError(err)

	s.NoError(r.Run(context.Background()), "happy path should succeed on fresh database")
	s.True(s.initialSchemaExists(), "initial schema tables should be created by migration")
	s.True(s.foreignKeyExists("FK_InvoiceItem_Invoice_InvoiceId"), "invoice item foreign key should exist")
	s.True(s.checkConstraintExists("CK_TYPE"), "transaction item type check constraint should exist")

	ver, dirty, verr := s.schemaVersion()
	s.NoError(verr)
	s.Equal(int64(1), ver, "schema_migrations should track version 1")
	s.False(dirty, "dirty flag must be false after successful migration")
}

// ---- Cenário 2: re-execução idempotente (RNF-01) ----

func (s *RunnerIntegrationSuite) TestScenario2_IdempotentRerun() {
	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "5m")
	s.T().Setenv("MIGRATION_BASELINE", "")

	r, err := migratorpkg.New(migratorpkg.WithLogger(discardLogger()))
	s.Require().NoError(err)

	s.Require().NoError(r.Run(context.Background()), "first run should succeed")
	s.NoError(r.Run(context.Background()), "second run (idempotent) should exit 0 via ErrNoChange")
	s.True(s.initialSchemaExists(), "initial schema tables must remain after idempotent re-run")
}

// ---- Cenário 3: pasta de migrations vazia — no-op legítimo (D-9) ----
// Injects a mock migrator that returns ErrMigrationsDirEmpty from Up().
// The runner must treat this as a no-op (exit 0) per D-9.
// Note: the real devkit-go migrator returns ErrMigrationFailed for empty FS;
// ErrMigrationsDirEmpty is the local sentinel that the runner handles as no-op.
// Using a mock factory isolates the D-9 handling logic cleanly.

func (s *RunnerIntegrationSuite) TestScenario3_EmptyMigrationsDir() {
	mockMig := migratorpkg_mocks.NewMockMigrator(s.T())
	mockMig.On("Up", mock.Anything).Return(migratorpkg.ErrMigrationsDirEmpty)

	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "5m")
	s.T().Setenv("MIGRATION_BASELINE", "")

	r, err := migratorpkg.New(
		migratorpkg.WithLogger(discardLogger()),
		migratorpkg.WithFactory(func(_ devkitmgr.Manager) (devkitmigration.Migrator, error) {
			return mockMig, nil
		}),
	)
	s.Require().NoError(err)

	s.NoError(r.Run(context.Background()), "runner with empty migrations dir must exit 0 (no-op, D-9)")
	s.False(s.initialSchemaExists(), "initial schema tables must NOT be created when migrations dir is empty")
}

// ---- Cenário 4: DSN inválido — D-53 (sem prefixo sqlserver://) ----
// Config validation rejects the DSN before any IO; no credentials appear in logs.

func (s *RunnerIntegrationSuite) TestScenario4_InvalidDSN() {
	scenarios := []struct {
		name string
		dsn  string
		want error
	}{
		{
			name: "wrong scheme mssql://",
			dsn:  "mssql://sa:S3cr3t@localhost:1433",
			want: migratorpkg.ErrInvalidDSNScheme,
		},
		{
			name: "no scheme",
			dsn:  "sa:S3cr3t@localhost:1433",
			want: migratorpkg.ErrInvalidDSNScheme,
		},
		{
			name: "empty DSN",
			dsn:  "",
			want: migratorpkg.ErrConfigMissingDSN,
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			s.T().Setenv("MSSQL_CONNECTION_STRING", sc.dsn)

			var logBuf bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

			_, err := migratorpkg.New(migratorpkg.WithLogger(log))
			s.Error(err, "New() must fail for invalid DSN %q", sc.dsn)
			s.True(errors.Is(err, sc.want), "error must wrap %v, got: %v", sc.want, err)

			// New() fails before any IO — no credential must appear in log output.
			s.NotContains(logBuf.String(), "S3cr3t", "password must not be logged")
		})
	}
}

// ---- Cenário 5: banco indisponível — retry esgota (RNF-05) ----
// 127.0.0.1:1 returns ECONNREFUSED immediately; with MIGRATION_TIMEOUT=4s the
// retry loop exhausts and returns an error wrapping ErrConnectionDeadline.

func (s *RunnerIntegrationSuite) TestScenario5_DBUnavailable() {
	s.T().Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:WrongPass1@127.0.0.1:1?database=master")
	s.T().Setenv("MIGRATION_TIMEOUT", "4s")
	s.T().Setenv("MIGRATION_BASELINE", "")

	var logBuf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	r, err := migratorpkg.New(migratorpkg.WithLogger(log))
	s.Require().NoError(err)

	runErr := r.Run(context.Background())
	s.Error(runErr, "runner must fail when DB is unavailable")

	// Error must indicate connection failure somewhere in the chain.
	s.True(
		errors.Is(runErr, migratorpkg.ErrConnectionDeadline) || containsAny(runErr.Error(), "open manager", "preflight", "connection"),
		"error must reflect connection failure, got: %v", runErr,
	)

	// Credentials must never appear in logs.
	s.NotContains(logBuf.String(), "WrongPass1", "password must not appear in logs")
}

// ---- Cenário 6: concorrência — lock advisory (RNF-09) ----
// Two goroutines run simultaneously against a fresh DB.
// One applies the migration; the other either succeeds (ErrNoChange) or
// fails with a lock conflict. Invariant: DB ends up in a consistent state.

func (s *RunnerIntegrationSuite) TestScenario6_ConcurrentRunners() {
	// Set env for both goroutines (process-wide, goroutines share env).
	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "5m")
	s.T().Setenv("MIGRATION_BASELINE", "")

	const n = 2
	errs := make([]error, n)

	var (
		wg    sync.WaitGroup
		ready = make(chan struct{}) // start gate: all goroutines wait until closed
	)
	close(ready)

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-ready

			r, newErr := migratorpkg.New(
				migratorpkg.WithLogger(discardLogger()),
				// Inject DSN explicitly so goroutines are independent of env reads.
				migratorpkg.WithOpenFunc(func(ctx context.Context, _ string, log *slog.Logger) (devkitmgr.Manager, error) {
					return migratorpkg.RetryOpen(ctx, s.dsn, log)
				}),
				migratorpkg.WithFactory(func(mgr devkitmgr.Manager) (devkitmigration.Migrator, error) {
					return devkitmigration.New(mgr,
						devkitmigration.EmbedFS{FS: localdatabase.MigrationsFS(), Root: "migrations"},
						devkitmigration.WithDSN(s.dsn),
					)
				}),
			)
			if newErr != nil {
				errs[i] = newErr
				return
			}
			errs[i] = r.Run(context.Background())
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, e := range errs {
		if e == nil {
			successCount++
		}
	}
	s.GreaterOrEqual(successCount, 1, "at least one concurrent runner must succeed; errors: %v", errs)
	s.True(s.initialSchemaExists(), "initial schema tables must exist after concurrent migrations")

	ver, dirty, verr := s.schemaVersion()
	s.NoError(verr)
	s.Equal(int64(1), ver, "schema_migrations must reflect version 1 after concurrent runs")
	s.False(dirty, "dirty flag must be false")
}

// ---- Cenário 7: SIGTERM gracioso — shutdown via ctx cancellation (RNF-12) ----
// A pre-cancelled context simulates SIGTERM. The runner must return an error
// without blocking.

func (s *RunnerIntegrationSuite) TestScenario7_SIGTERMGraceful() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // simulate SIGTERM: cancel before Run

	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "30s")
	s.T().Setenv("MIGRATION_BASELINE", "")

	r, err := migratorpkg.New(migratorpkg.WithLogger(discardLogger()))
	s.Require().NoError(err)

	start := time.Now()
	runErr := r.Run(ctx)
	elapsed := time.Since(start)

	s.Error(runErr, "runner must fail when context is cancelled (SIGTERM simulation)")
	s.Less(elapsed, 30*time.Second, "runner must not block after context cancellation; elapsed: %v", elapsed)
}

// ---- Cenário 8: baseline + Force — ambientes preexistentes (D-42) ----
// Creates the full initial schema manually (pre-existing DB), forces version 1 via
// MIGRATION_BASELINE=1, then verifies Up() is a no-op (ErrNoChange → exit 0).

func (s *RunnerIntegrationSuite) TestScenario8_BaselineForce() {
	ctx := context.Background()

	ddl, err := fs.ReadFile(localdatabase.MigrationsFS(), "migrations/000001_initial_schema.up.sql")
	s.Require().NoError(err, "read embedded initial schema migration")

	// Pre-create the schema to simulate a DB that already has version 1.
	_, err = s.db.ExecContext(ctx, string(ddl))
	s.Require().NoError(err, "pre-create initial schema")
	s.True(s.initialSchemaExists(), "initial schema should exist before baseline force")

	realFactory := func(mgr devkitmgr.Manager) (devkitmigration.Migrator, error) {
		return devkitmigration.New(mgr,
			devkitmigration.EmbedFS{FS: localdatabase.MigrationsFS(), Root: "migrations"},
			devkitmigration.WithDSN(s.dsn),
		)
	}

	// Step 1: Force baseline version 1 (marks 0001 as applied without running SQL).
	s.T().Setenv("MSSQL_CONNECTION_STRING", s.dsn)
	s.T().Setenv("MIGRATION_TIMEOUT", "5m")
	s.T().Setenv("MIGRATION_BASELINE", "1")

	r, err := migratorpkg.New(
		migratorpkg.WithLogger(discardLogger()),
		migratorpkg.WithFactory(realFactory),
	)
	s.Require().NoError(err)
	s.Require().NoError(r.Run(ctx), "baseline Force(1) should succeed")

	// Step 2: verify schema_migrations records version 1 (not dirty).
	ver, dirty, verr := s.schemaVersion()
	s.NoError(verr, "schema_migrations should exist after Force")
	s.Equal(int64(1), ver, "version must be 1 after Force(1)")
	s.False(dirty, "dirty must be false after Force(1)")

	// Step 3: run Up without baseline — version 1 is already applied → ErrNoChange → exit 0.
	s.T().Setenv("MIGRATION_BASELINE", "")

	r2, err := migratorpkg.New(
		migratorpkg.WithLogger(discardLogger()),
		migratorpkg.WithFactory(realFactory),
	)
	s.Require().NoError(err)
	s.NoError(r2.Run(ctx), "Up() after baseline Force must be no-op (ErrNoChange → exit 0)")
	s.True(s.initialSchemaExists(), "initial schema tables must still exist")
}

// ---- Helpers ----

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *RunnerIntegrationSuite) initialSchemaExists() bool {
	for _, tableName := range initialSchemaTableNames {
		if !s.tableExists(tableName) {
			return false
		}
	}
	return true
}

func (s *RunnerIntegrationSuite) tableExists(tableName string) bool {
	var cnt int
	_ = s.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = @tableName
	`, sql.Named("tableName", tableName)).Scan(&cnt)
	return cnt == 1
}

func (s *RunnerIntegrationSuite) foreignKeyExists(name string) bool {
	var cnt int
	_ = s.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*)
		FROM sys.foreign_keys
		WHERE name = @name
	`, sql.Named("name", name)).Scan(&cnt)
	return cnt == 1
}

func (s *RunnerIntegrationSuite) checkConstraintExists(name string) bool {
	var cnt int
	_ = s.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*)
		FROM sys.check_constraints
		WHERE name = @name
	`, sql.Named("name", name)).Scan(&cnt)
	return cnt == 1
}

func (s *RunnerIntegrationSuite) schemaVersion() (int64, bool, error) {
	var version int64
	var dirty bool
	err := s.db.QueryRowContext(context.Background(),
		"SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	return version, dirty, err
}

// containsAny reports whether s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}

var initialSchemaTableNames = []string{
	"Bill",
	"Category",
	"Flag",
	"User",
	"BillItem",
	"Card",
	"Invoice",
	"InvoiceItem",
	"Transaction",
	"TransactionItem",
}

var cleanupStatements = []string{
	"IF OBJECT_ID('dbo.TransactionItem', 'U') IS NOT NULL DROP TABLE dbo.[TransactionItem]",
	"IF OBJECT_ID('dbo.InvoiceItem', 'U') IS NOT NULL DROP TABLE dbo.[InvoiceItem]",
	"IF OBJECT_ID('dbo.Invoice', 'U') IS NOT NULL DROP TABLE dbo.[Invoice]",
	"IF OBJECT_ID('dbo.Card', 'U') IS NOT NULL DROP TABLE dbo.[Card]",
	"IF OBJECT_ID('dbo.BillItem', 'U') IS NOT NULL DROP TABLE dbo.[BillItem]",
	"IF OBJECT_ID('dbo.[Transaction]', 'U') IS NOT NULL DROP TABLE dbo.[Transaction]",
	"IF OBJECT_ID('dbo.Bill', 'U') IS NOT NULL DROP TABLE dbo.[Bill]",
	"IF OBJECT_ID('dbo.Category', 'U') IS NOT NULL DROP TABLE dbo.[Category]",
	"IF OBJECT_ID('dbo.Flag', 'U') IS NOT NULL DROP TABLE dbo.[Flag]",
	"IF OBJECT_ID('dbo.[User]', 'U') IS NOT NULL DROP TABLE dbo.[User]",
	"IF OBJECT_ID('schema_migrations', 'U') IS NOT NULL DROP TABLE schema_migrations",
}
