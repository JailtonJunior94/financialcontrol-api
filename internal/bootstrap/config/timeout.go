package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const defaultShutdownTimeout = 15 * time.Second

var ErrInvalidShutdownTimeout = errors.New("HTTP_SHUTDOWN_TIMEOUT must be a positive duration")

// ShutdownTimeout holds a validated graceful-shutdown duration read from HTTP_SHUTDOWN_TIMEOUT.
type ShutdownTimeout struct {
	d time.Duration
}

// NewShutdownTimeout reads HTTP_SHUTDOWN_TIMEOUT and validates it.
// Returns 15s default when env is absent. Fails with ErrInvalidShutdownTimeout when
// the env is present but empty, unparseable, or <= 0.
func NewShutdownTimeout() (ShutdownTimeout, error) {
	raw, ok := os.LookupEnv("HTTP_SHUTDOWN_TIMEOUT")
	if !ok {
		return ShutdownTimeout{d: defaultShutdownTimeout}, nil
	}

	if raw == "" {
		return ShutdownTimeout{}, fmt.Errorf("bootstrap config: %w", ErrInvalidShutdownTimeout)
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return ShutdownTimeout{}, fmt.Errorf("bootstrap config: %w", ErrInvalidShutdownTimeout)
	}

	return ShutdownTimeout{d: d}, nil
}

func (t ShutdownTimeout) Duration() time.Duration { return t.d }
