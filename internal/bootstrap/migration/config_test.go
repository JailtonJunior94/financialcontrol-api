package migration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		envs         map[string]string
		wantErr      error
		wantErrStr   string
		wantTimeout  time.Duration
		wantBaseline uint
	}{
		{
			name:    "missing DSN returns ErrConfigMissingDSN",
			envs:    map[string]string{},
			wantErr: ErrConfigMissingDSN,
		},
		{
			name:    "wrong DSN scheme returns ErrInvalidDSNScheme",
			envs:    map[string]string{"MSSQL_CONNECTION_STRING": "postgres://host"},
			wantErr: ErrInvalidDSNScheme,
		},
		{
			name:        "valid DSN uses default timeout",
			envs:        map[string]string{"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost"},
			wantTimeout: 10 * time.Minute,
		},
		{
			name: "custom timeout parsed",
			envs: map[string]string{
				"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost",
				"MIGRATION_TIMEOUT":       "5m",
			},
			wantTimeout: 5 * time.Minute,
		},
		{
			name: "invalid timeout returns error",
			envs: map[string]string{
				"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost",
				"MIGRATION_TIMEOUT":       "notaduration",
			},
			wantErrStr: "invalid MIGRATION_TIMEOUT",
		},
		{
			name:    "baseline=0 returns ErrInvalidBaselineVersion",
			envs:    map[string]string{"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost", "MIGRATION_BASELINE": "0"},
			wantErr: ErrInvalidBaselineVersion,
		},
		{
			name:    "non-numeric baseline returns ErrInvalidBaselineVersion",
			envs:    map[string]string{"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost", "MIGRATION_BASELINE": "abc"},
			wantErr: ErrInvalidBaselineVersion,
		},
		{
			name: "valid baseline parsed",
			envs: map[string]string{
				"MSSQL_CONNECTION_STRING": "sqlserver://sa:pass@localhost",
				"MIGRATION_BASELINE":      "3",
			},
			wantTimeout:  10 * time.Minute,
			wantBaseline: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MSSQL_CONNECTION_STRING", "")
			t.Setenv("MIGRATION_TIMEOUT", "")
			t.Setenv("MIGRATION_BASELINE", "")
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			cfg, err := LoadConfig()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			if tc.wantErrStr != "" {
				require.ErrorContains(t, err, tc.wantErrStr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.envs["MSSQL_CONNECTION_STRING"], cfg.DSN)
			assert.Equal(t, tc.wantTimeout, cfg.Timeout)
			assert.Equal(t, tc.wantBaseline, cfg.BaselineVersion)
		})
	}
}
