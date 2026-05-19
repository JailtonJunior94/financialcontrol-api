package config_test

import (
	"errors"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceIdentity(t *testing.T) {
	tests := []struct {
		name        string
		envName     string
		envVersion  string
		envEnv      string
		wantName    string
		wantVersion string
		wantEnv     string
		wantErr     error
	}{
		{
			name:        "valid development",
			envName:     "my-service",
			envVersion:  "1.0.0",
			envEnv:      "development",
			wantName:    "my-service",
			wantVersion: "1.0.0",
			wantEnv:     "development",
		},
		{
			name:        "valid staging",
			envName:     "my-service",
			envVersion:  "1.0.0",
			envEnv:      "staging",
			wantName:    "my-service",
			wantVersion: "1.0.0",
			wantEnv:     "staging",
		},
		{
			name:        "valid production",
			envName:     "my-service",
			envVersion:  "1.0.0",
			envEnv:      "production",
			wantName:    "my-service",
			wantVersion: "1.0.0",
			wantEnv:     "production",
		},
		{
			name:       "empty SERVICE_NAME",
			envName:    "",
			envVersion: "1.0.0",
			envEnv:     "development",
			wantErr:    config.ErrServiceNameRequired,
		},
		{
			name:       "empty SERVICE_VERSION",
			envName:    "my-service",
			envVersion: "",
			envEnv:     "development",
			wantErr:    config.ErrServiceVersionRequired,
		},
		{
			name:       "empty ENVIRONMENT",
			envName:    "my-service",
			envVersion: "1.0.0",
			envEnv:     "",
			wantErr:    config.ErrInvalidEnvironment,
		},
		{
			name:       "ENVIRONMENT=Production (capitalized)",
			envName:    "my-service",
			envVersion: "1.0.0",
			envEnv:     "Production",
			wantErr:    config.ErrInvalidEnvironment,
		},
		{
			name:       "ENVIRONMENT=PROD",
			envName:    "my-service",
			envVersion: "1.0.0",
			envEnv:     "PROD",
			wantErr:    config.ErrInvalidEnvironment,
		},
		{
			name:       "ENVIRONMENT=dev",
			envName:    "my-service",
			envVersion: "1.0.0",
			envEnv:     "dev",
			wantErr:    config.ErrInvalidEnvironment,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SERVICE_NAME", tc.envName)
			t.Setenv("SERVICE_VERSION", tc.envVersion)
			t.Setenv("ENVIRONMENT", tc.envEnv)

			got, err := config.NewServiceIdentity()

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, got.Name())
			assert.Equal(t, tc.wantVersion, got.Version())
			assert.Equal(t, tc.wantEnv, got.Environment())
		})
	}
}
