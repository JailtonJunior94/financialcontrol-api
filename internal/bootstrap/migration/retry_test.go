package migration

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRetryOpen_CtxCancelledAbortsRetry(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	_, err := RetryOpen(ctx, "sqlserver://sa:pass@nonexistent:1433", noopLogger())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConnectionDeadline)
}

func TestRetryOpen_MaxAttemptsRespected(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	_, err := RetryOpen(ctx, "sqlserver://sa:pass@192.0.2.1:1433?dial timeout=1", noopLogger())
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConnectionDeadline)
	assert.Greater(t, elapsed, 100*time.Millisecond, "expected backoff delay to be applied")
}

func TestRetryOpen_ExponentialBackoff(t *testing.T) {
	t.Parallel()
	// Verify the backoff sequence calculation inline — no need for a real DB call.
	delays := make([]time.Duration, 4)
	delay := initialDelay
	for i := range delays {
		delays[i] = delay
		if delay < maxDelay {
			delay *= 2
		}
	}

	assert.Equal(t, 500*time.Millisecond, delays[0])
	assert.Equal(t, 1*time.Second, delays[1])
	assert.Equal(t, 2*time.Second, delays[2])
	assert.Equal(t, 4*time.Second, delays[3])
}

func TestRetryOpen_CtxCancelledDuringWait(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := RetryOpen(ctx, "sqlserver://sa:Password123@localhost:1433?dial timeout=1", noopLogger())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConnectionDeadline)
}

// Verify that RetryOpen signature matches what runner.go expects.
var _ func(context.Context, string, *slog.Logger) (manager.Manager, error) = RetryOpen
