package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/database/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/mocks"
)

func buildTestRunner(
	t *testing.T,
	envs map[string]string,
	mockMgr *mocks.MockManager,
	mockMig *mocks.MockMigrator,
) Runner {
	t.Helper()
	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")
	t.Setenv("MIGRATION_BASELINE", "")
	t.Setenv("MIGRATION_TIMEOUT", "")
	for k, v := range envs {
		t.Setenv(k, v)
	}
	openFn := func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) {
		return mockMgr, nil
	}
	factoryFn := func(_ manager.Manager) (migration.Migrator, error) {
		return mockMig, nil
	}
	r, err := New(
		WithOpenFunc(openFn),
		WithFactory(factoryFn),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	require.NoError(t, err)
	return r
}

func TestRunner_HappyPath(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Return(nil)

	r := buildTestRunner(t, nil, mockMgr, mockMig)
	require.NoError(t, r.Run(context.Background()))
}

func TestRunner_NoChange(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Return(migration.ErrNoChange)

	r := buildTestRunner(t, nil, mockMgr, mockMig)
	require.NoError(t, r.Run(context.Background()))
}

func TestRunner_MigrationsDirEmpty_IsNoOp(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Return(ErrMigrationsDirEmpty)

	r := buildTestRunner(t, nil, mockMgr, mockMig)
	require.NoError(t, r.Run(context.Background()))
}

func TestRunner_UpError_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	upErr := errors.New("migration failed")
	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Return(upErr)

	r := buildTestRunner(t, nil, mockMgr, mockMig)
	err := r.Run(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "migrate up")
}

func TestRunner_PreflightFails_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	pingErr := errors.New("ping failed")
	mockMgr.On("Ping", mock.Anything).Return(pingErr)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	r := buildTestRunner(t, nil, mockMgr, mockMig)
	err := r.Run(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "preflight")
}

func TestRunner_BaselineVersion_CallsForce(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Force", mock.Anything, 5).Return(nil)

	r := buildTestRunner(t, map[string]string{"MIGRATION_BASELINE": "5"}, mockMgr, mockMig)
	require.NoError(t, r.Run(context.Background()))
}

func TestRunner_OpenFails_ReturnsError(t *testing.T) {
	openErr := errors.New("open failed")
	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")

	r, err := New(
		WithOpenFunc(func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) {
			return nil, openErr
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	require.NoError(t, err)
	runErr := r.Run(context.Background())
	require.Error(t, runErr)
	assert.ErrorContains(t, runErr, "open manager")
}

func TestRunner_FactoryFails_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)

	factoryErr := errors.New("factory failed")
	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")
	r, err := New(
		WithOpenFunc(func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) {
			return mockMgr, nil
		}),
		WithFactory(func(_ manager.Manager) (migration.Migrator, error) {
			return nil, factoryErr
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	require.NoError(t, err)
	runErr := r.Run(context.Background())
	require.Error(t, runErr)
	assert.ErrorContains(t, runErr, "build migrator")
}

func TestRunner_TimeoutExpires_ReturnsDeadlineError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	// Open succeeds, ping succeeds, but Up blocks until context times out.
	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(context.DeadlineExceeded)

	r := buildTestRunner(t, map[string]string{"MIGRATION_TIMEOUT": "50ms"}, mockMgr, mockMig)
	start := time.Now()
	err := r.Run(context.Background())
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second, "expected runner to respect the configured timeout")
}

func TestRunner_LogsAreJSON(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)
	mockMig.On("Up", mock.Anything).Return(nil)

	var buf bytes.Buffer
	log := slog.New(&redactingHandler{
		inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
	})

	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")
	r, err := New(
		WithOpenFunc(func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) { return mockMgr, nil }),
		WithFactory(func(_ manager.Manager) (migration.Migrator, error) { return mockMig, nil }),
		WithLogger(log),
	)
	require.NoError(t, err)
	require.NoError(t, r.Run(context.Background()))

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var record map[string]any
		assert.NoError(t, json.Unmarshal(line, &record), "log line is not valid JSON: %s", line)
	}
}

func TestNew_MissingDSN_ReturnsError(t *testing.T) {
	t.Setenv("MSSQL_CONNECTION_STRING", "")
	_, err := New()
	require.ErrorIs(t, err, ErrConfigMissingDSN)
}

// ──────────────────────────────────────────────────────────────────────────────
// Smoke pipeline tests (applyWithSmoke)
// ──────────────────────────────────────────────────────────────────────────────

func buildSmokeRunner(
	t *testing.T,
	mockMgr *mocks.MockManager,
	mockMig *mocks.MockMigrator,
	hookFn SmokeHookFn,
) Runner {
	t.Helper()
	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")
	t.Setenv("MIGRATION_BASELINE", "")
	t.Setenv("MIGRATION_TIMEOUT", "")

	r, err := New(
		WithOpenFunc(func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) {
			return mockMgr, nil
		}),
		WithFactory(func(_ manager.Manager) (migration.Migrator, error) {
			return mockMig, nil
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		withTestHookFn("finance", hookFn),
	)
	require.NoError(t, err)
	return r
}

func TestRunner_Smoke_GoldenPath_DirtyAtV3(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	hookCalled := false
	hookFn := func(_ context.Context, _ *slog.Logger) error {
		hookCalled = true
		return nil
	}

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	// Phase 1: Up returns an error (gate fires at v3 dirty).
	mockMig.On("Up", mock.Anything).Return(errors.New("dirty migration at version 3")).Once()
	// Version check: v3 dirty.
	mockMig.On("Version", mock.Anything).Return(uint(3), true, nil).Once()
	// Force reset to v2.
	mockMig.On("Force", mock.Anything, smokeTriggerVersion).Return(nil).Once()
	// Phase 3: Up succeeds.
	mockMig.On("Up", mock.Anything).Return(nil).Once()

	r := buildSmokeRunner(t, mockMgr, mockMig, hookFn)
	require.NoError(t, r.Run(context.Background()))
	assert.True(t, hookCalled, "smoke hook must be invoked")
}

func TestRunner_Smoke_AlreadyComplete_SkipsHook(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	hookCalled := false
	hookFn := func(_ context.Context, _ *slog.Logger) error {
		hookCalled = true
		return nil
	}

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	// Phase 1: Up returns no change (all migrations already applied).
	mockMig.On("Up", mock.Anything).Return(migration.ErrNoChange).Once()
	// Version: v3, not dirty — beyond trigger version, already done.
	mockMig.On("Version", mock.Anything).Return(uint(3), false, nil).Once()

	r := buildSmokeRunner(t, mockMgr, mockMig, hookFn)
	require.NoError(t, r.Run(context.Background()))
	assert.False(t, hookCalled, "smoke hook must NOT be invoked when migrations are complete")
}

func TestRunner_Smoke_HookFails_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	hookErr := errors.New("smoke probe failed")
	hookFn := func(_ context.Context, _ *slog.Logger) error { return hookErr }

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	mockMig.On("Up", mock.Anything).Return(errors.New("gate fired")).Once()
	mockMig.On("Version", mock.Anything).Return(uint(3), true, nil).Once()
	mockMig.On("Force", mock.Anything, smokeTriggerVersion).Return(nil).Once()

	r := buildSmokeRunner(t, mockMgr, mockMig, hookFn)
	err := r.Run(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, hookErr)
	assert.ErrorContains(t, err, "smoke")
}

func TestRunner_Smoke_StalledBelowTrigger_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	hookFn := func(_ context.Context, _ *slog.Logger) error { return nil }

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	mockMig.On("Up", mock.Anything).Return(errors.New("failed at v1")).Once()
	// Version 1, dirty — stalled before trigger version.
	mockMig.On("Version", mock.Anything).Return(uint(1), true, nil).Once()

	r := buildSmokeRunner(t, mockMgr, mockMig, hookFn)
	err := r.Run(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "migration stalled")
}

func TestRunner_Smoke_UnregisteredHook_ReturnsError(t *testing.T) {
	mockMgr := mocks.NewMockManager(t)
	mockMig := mocks.NewMockMigrator(t)

	mockMgr.On("Ping", mock.Anything).Return(nil)
	mockMgr.On("Shutdown", mock.Anything).Return(nil)

	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://sa:pass@localhost")
	t.Setenv("MIGRATION_BASELINE", "")
	t.Setenv("MIGRATION_TIMEOUT", "")

	r, err := New(
		WithOpenFunc(func(_ context.Context, _ string, _ *slog.Logger) (manager.Manager, error) {
			return mockMgr, nil
		}),
		WithFactory(func(_ manager.Manager) (migration.Migrator, error) {
			return mockMig, nil
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithSmokeHook("nonexistent"),
	)
	require.NoError(t, err)

	// applyWithSmoke checks globalHooks before calling Up, so no migrator calls are expected.
	err = r.Run(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "not registered")
}
