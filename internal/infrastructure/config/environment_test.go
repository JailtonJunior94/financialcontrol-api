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
		expectedDSN string
	}{
		{
			name:        "development",
			environment: "Development",
			expectedDSN: "sqlserver://sa:@docker@2022@localhost:1433?database=FinancialControlDB",
		},
		{
			name:        "staging",
			environment: "Staging",
			expectedDSN: "sqlserver://sa:@docker@2021@mssql?database=FinancialControlDB",
		},
		{
			name:        "production",
			environment: "Production",
			expectedDSN: "sqlserver://DB_A453C8_FinancialControl_admin:@stefany@1994@SQL5053.site4now.net?database=DB_A453C8_FinancialControl",
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			resetGlobals()
			t.Cleanup(resetGlobals)
			t.Setenv("ENVIRONMENT", testCase.environment)

			err := Load()

			require.NoError(t, err)
			require.Equal(t, testCase.environment, Environment)
			require.Equal(t, testCase.expectedDSN, SqlConnectionString)
			require.NotEmpty(t, JwtSecret)
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
