package instrumented

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureConnectionString(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		env         string
		wantContain string
		wantAbsent  string
		wantErr     error
	}{
		{
			name:        "injects app name when absent",
			raw:         "server=localhost;database=mydb;user id=sa;password=pass",
			env:         "production",
			wantContain: "app name=financialcontrol-api-production",
		},
		{
			name:        "overwrites pre-existing app name",
			raw:         "server=x;app name=other;database=mydb",
			env:         "production",
			wantContain: "app name=financialcontrol-api-production",
			wantAbsent:  "other",
		},
		{
			name:        "case-insensitive key match",
			raw:         "server=x;APP NAME=legacy_tool",
			env:         "staging",
			wantContain: "app name=financialcontrol-api-staging",
			wantAbsent:  "legacy_tool",
		},
		{
			name:        "development environment",
			raw:         "server=localhost",
			env:         "development",
			wantContain: "financialcontrol-api-development",
		},
		{
			name:    "empty raw returns error",
			raw:     "",
			env:     "production",
			wantErr: ErrInvalidConnectionString,
		},
		{
			name:    "empty env returns error",
			raw:     "server=localhost",
			env:     "",
			wantErr: ErrInvalidConnectionString,
		},
		{
			name:        "trailing semicolon is handled",
			raw:         "server=localhost;",
			env:         "staging",
			wantContain: "financialcontrol-api-staging",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConfigureConnectionString(tc.raw, tc.env)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr),
					"expected error %v, got %v", tc.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, got, tc.wantContain)
			if tc.wantAbsent != "" {
				assert.NotContains(t, got, tc.wantAbsent)
			}
		})
	}
}

func TestConfigureConnectionString_Idempotent(t *testing.T) {
	raw := "server=localhost;database=mydb"
	env := "production"

	first, err := ConfigureConnectionString(raw, env)
	require.NoError(t, err)

	second, err := ConfigureConnectionString(first, env)
	require.NoError(t, err)

	third, err := ConfigureConnectionString(second, env)
	require.NoError(t, err)

	assert.Equal(t, first, second, "second call must equal first")
	assert.Equal(t, second, third, "third call must equal second")
}

func TestConfigureConnectionString_OverwriteWithDifferentEnv(t *testing.T) {
	raw := "server=localhost;app name=financialcontrol-api-development"

	got, err := ConfigureConnectionString(raw, "production")
	require.NoError(t, err)

	assert.Contains(t, got, "financialcontrol-api-production")
	assert.NotContains(t, got, "development")
}
