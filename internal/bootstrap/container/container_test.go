package container_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/noop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapconfig "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	migrationmocks "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/mocks"
	bootstrapobs "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

// spyObservability wraps a noop provider and counts Shutdown invocations. Used to
// prove BuildRuntime rolls back observability on downstream failure.
type spyObservability struct {
	observability.Observability
	shutdowns int32
}

func newSpyObservability() *spyObservability {
	return &spyObservability{Observability: noop.NewProvider()}
}

func (s *spyObservability) Shutdown(ctx context.Context) error {
	atomic.AddInt32(&s.shutdowns, 1)
	return s.Observability.Shutdown(ctx)
}

func (s *spyObservability) ShutdownCount() int32 { return atomic.LoadInt32(&s.shutdowns) }

// setJwtSecret overrides the global config.JwtSecret for the duration of t and restores it on cleanup.
func setJwtSecret(t *testing.T, secret string) {
	t.Helper()
	original := config.JwtSecret
	t.Cleanup(func() { config.JwtSecret = original })
	config.JwtSecret = secret
}

// setExpiration overrides config.ExpirationAt for the duration of t and restores it on cleanup.
func setExpiration(t *testing.T, hours int) {
	t.Helper()
	original := config.ExpirationAt
	t.Cleanup(func() { config.ExpirationAt = original })
	config.ExpirationAt = hours
}

// TestBuild_Happy verifies that Build wires all modules without error given valid inputs.
func TestBuild_Happy(t *testing.T) {
	setJwtSecret(t, "super-secret-jwt-key-for-unit-tests!")
	setExpiration(t, 1)

	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	t.Setenv("ENVIRONMENT", "development")

	id, err := bootstrapconfig.NewServiceIdentity()
	require.NoError(t, err)

	timeout, err := bootstrapconfig.NewShutdownTimeout()
	require.NoError(t, err)

	mockMgr := migrationmocks.NewMockManager(t)
	mockMgr.On("DBTX", mock.Anything).Return(nil)

	obs := noop.NewProvider()

	c, err := bootstrapcontainer.Build(context.Background(), mockMgr, obs, id, timeout)

	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, mockMgr, c.DBManager)
	assert.Equal(t, obs, c.Observability)
	assert.Equal(t, "test-svc", c.Identity.Name())
	assert.Equal(t, "1.0.0", c.Identity.Version())
	assert.Equal(t, "development", c.Identity.Environment())
	assert.NotNil(t, c.JwtParser)
	assert.NotNil(t, c.IdentityModule)
	assert.NotNil(t, c.CardsModule)
	assert.NotNil(t, c.CategoriesModule)
	assert.NotNil(t, c.FinanceModule)
	assert.NotNil(t, c.BusinessMetrics, "BusinessMetrics must be wired so finance module receives a real recorder")
}

// TestBuildRuntime_FailsOnMissingServiceIdentity verifies that BuildRuntime propagates
// identity errors (SERVICE_NAME required) as bootstrap container errors.
// Uses testdata/ so config.Load finds a lowercase ENVIRONMENT config file.
func TestBuildRuntime_FailsOnMissingServiceIdentity(t *testing.T) {
	t.Chdir("testdata")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("SERVICE_NAME", "")

	c, err := bootstrapcontainer.BuildRuntime(context.Background())

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, bootstrapconfig.ErrServiceNameRequired),
		"expected ErrServiceNameRequired in chain, got: %v", err)
}

// TestBuildRuntime_FailsOnInvalidOtelExporter verifies that BuildRuntime propagates
// observability settings errors when the OTLP protocol value is unknown (ADR-005).
// Uses testdata/ so config.Load finds a lowercase ENVIRONMENT config file.
func TestBuildRuntime_FailsOnInvalidOtelExporter(t *testing.T) {
	t.Chdir("testdata")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "invalid-protocol")

	c, err := bootstrapcontainer.BuildRuntime(context.Background())

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, bootstrapobs.ErrInvalidOTLPProtocol),
		"expected ErrInvalidOTLPProtocol in chain, got: %v", err)
}

// TestBuildRuntime_FailsOnMissingDSN verifies that BuildRuntime propagates the database
// error when the connection string is absent (simulates MSSQL unavailable at startup).
// Uses testdata/ so config.Load finds a lowercase ENVIRONMENT config file.
func TestBuildRuntime_FailsOnMissingDSN(t *testing.T) {
	t.Chdir("testdata")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	c, err := bootstrapcontainer.BuildRuntime(context.Background())

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, pkgdatabase.ErrConfigMissingDSN),
		"expected ErrConfigMissingDSN in chain, got: %v", err)
}

// TestBuildRuntime_ShutsDownObservabilityOnDBFailure verifies the rollback path:
// when database.OpenManager fails after observability has been constructed,
// BuildRuntime must call obs.Shutdown so OTel provider goroutines do not leak.
// Regression for the leak documented in tasks/prd-migration-devkit-httpserver/bugfix_report.md.
func TestBuildRuntime_ShutsDownObservabilityOnDBFailure(t *testing.T) {
	t.Chdir("testdata")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	// MSSQL_CONNECTION_STRING intentionally unset → ErrConfigMissingDSN.

	spy := newSpyObservability()
	restore := bootstrapcontainer.SetNewObservabilityForTest(
		func(_ context.Context, _ bootstrapobs.Settings) (observability.Observability, error) {
			return spy, nil
		},
	)
	t.Cleanup(restore)

	c, err := bootstrapcontainer.BuildRuntime(context.Background())

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, pkgdatabase.ErrConfigMissingDSN))
	assert.Equal(t, int32(1), spy.ShutdownCount(),
		"obs.Shutdown must be called exactly once when DB step fails")
}

// TestBuild_FailsOnEmptyJWTSecret verifies that Build propagates ErrSecretTooShort
// when config.JwtSecret is empty. DBTX is never reached so no expectation is set.
func TestBuild_FailsOnEmptyJWTSecret(t *testing.T) {
	setJwtSecret(t, "")

	mockMgr := migrationmocks.NewMockManager(t)
	obs := noop.NewProvider()
	var id bootstrapconfig.ServiceIdentity
	var timeout bootstrapconfig.ShutdownTimeout

	c, err := bootstrapcontainer.Build(context.Background(), mockMgr, obs, id, timeout)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.True(t, errors.Is(err, pkgjwt.ErrSecretTooShort),
		"expected ErrSecretTooShort, got: %v", err)
}
