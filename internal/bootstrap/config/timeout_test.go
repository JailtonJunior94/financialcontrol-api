package config_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShutdownTimeout(t *testing.T) {
	tests := []struct {
		name     string
		absent   bool   // true: ensure env is unset; false: set env to envValue
		envValue string
		want     time.Duration
		wantErr  error
	}{
		{
			name:   "default 15s when env absent",
			absent: true,
			want:   15 * time.Second,
		},
		{
			name:     "explicit 15s",
			envValue: "15s",
			want:     15 * time.Second,
		},
		{
			name:     "explicit 1m",
			envValue: "1m",
			want:     time.Minute,
		},
		{
			name:     "empty string with env present",
			envValue: "",
			wantErr:  config.ErrInvalidShutdownTimeout,
		},
		{
			name:     "invalid abc",
			envValue: "abc",
			wantErr:  config.ErrInvalidShutdownTimeout,
		},
		{
			name:     "zero value 0s",
			envValue: "0s",
			wantErr:  config.ErrInvalidShutdownTimeout,
		},
		{
			name:     "negative -1s",
			envValue: "-1s",
			wantErr:  config.ErrInvalidShutdownTimeout,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.absent {
				orig, wasSet := os.LookupEnv("HTTP_SHUTDOWN_TIMEOUT")
				require.NoError(t, os.Unsetenv("HTTP_SHUTDOWN_TIMEOUT"))
				t.Cleanup(func() {
					if wasSet {
						_ = os.Setenv("HTTP_SHUTDOWN_TIMEOUT", orig)
					} else {
						_ = os.Unsetenv("HTTP_SHUTDOWN_TIMEOUT")
					}
				})
			} else {
				t.Setenv("HTTP_SHUTDOWN_TIMEOUT", tc.envValue)
			}

			got, err := config.NewShutdownTimeout()

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Duration())
		})
	}
}
