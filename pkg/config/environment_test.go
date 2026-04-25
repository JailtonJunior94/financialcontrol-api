package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"Development": "config.Development.yaml",
		"Staging":     "config.Staging.yaml",
		"Production":  "config.Production.yaml",
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
		{name: "development", environment: "Development"},
		{name: "staging", environment: "Staging"},
		{name: "production", environment: "Production"},
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

func resetGlobals() {
	Environment = ""
	SqlConnectionString = ""
	JwtSecret = ""
	ExpirationAt = 0
}
