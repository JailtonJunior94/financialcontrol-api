package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"development": "config.development.yaml",
		"staging":     "config.staging.yaml",
		"production":  "config.production.yaml",
	}

	for environment, expectedFile := range testCases {

		t.Run(environment, func(t *testing.T) {
			t.Parallel()

			path, err := resolveConfigPath(environment)

			require.NoError(t, err)
			require.FileExists(t, path)
			require.Equal(t, expectedFile, filepath.Base(path))
			require.Equal(t, configDir, filepath.Base(filepath.Dir(path)))
		})
	}
}

func TestLoadReadsEnvironmentConfigFromConfigsDirectory(t *testing.T) {
	testCases := []struct {
		name        string
		environment string
	}{
		{name: "development", environment: "development"},
		{name: "staging", environment: "staging"},
		{name: "production", environment: "production"},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			resetGlobals()
			t.Cleanup(resetGlobals)
			t.Setenv("ENVIRONMENT", testCase.environment)
			t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://test:pass@localhost:1433?database=TestDB")
			t.Setenv("JWT_SECRET", "dGVzdC1zZWNyZXQ=")

			err := Load()

			require.NoError(t, err)
			require.Equal(t, testCase.environment, Environment)
			require.Equal(t, "sqlserver://test:pass@localhost:1433?database=TestDB", SqlConnectionString)
			require.Equal(t, "dGVzdC1zZWNyZXQ=", JwtSecret)
			require.Equal(t, 1, ExpirationAt)
		})
	}
}

func TestLoadReturnsExplicitErrorWhenEnvironmentIsMissing(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)
	t.Setenv("ENVIRONMENT", "")

	err := Load()

	require.Error(t, err)
	require.EqualError(t, err, "environment variable ENVIRONMENT is required")
}

func TestLoadReturnsExplicitErrorWhenConfigFileIsMissing(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)
	t.Setenv("ENVIRONMENT", "Missing")

	err := Load()

	require.Error(t, err)
	require.ErrorContains(t, err, `config.Missing.yaml`)
	require.ErrorContains(t, err, `configs`)
}

// TestLoad_DoesNotInterfereWithBootstrapEnvs documents that the bootstrap-layer envs
// (SERVICE_NAME, SERVICE_VERSION, HTTP_SHUTDOWN_TIMEOUT, OTEL_EXPORTER_OTLP_PROTOCOL,
// OTEL_EXPORTER_OTLP_ENDPOINT) are independent of pkg/config.Load.
// They are validated in internal/bootstrap/config and internal/bootstrap/observability.
func TestLoad_DoesNotInterfereWithBootstrapEnvs(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("MSSQL_CONNECTION_STRING", "sqlserver://test:pass@localhost:1433?database=TestDB")
	t.Setenv("JWT_SECRET", "dGVzdC1zZWNyZXQ=")
	// Bootstrap-only envs coexisting alongside pkg/config envs.
	t.Setenv("SERVICE_NAME", "test-svc")
	t.Setenv("SERVICE_VERSION", "1.0.0")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "30s")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	err := Load()

	require.NoError(t, err, "bootstrap env vars must not interfere with pkg/config.Load")
	require.Equal(t, "development", Environment)
}

func resetGlobals() {
	Environment = ""
	SqlConnectionString = ""
	JwtSecret = ""
	ExpirationAt = 0
}
