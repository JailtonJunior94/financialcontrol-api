package migration

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	localdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
)

const (
	initialDelay = 500 * time.Millisecond
	maxDelay     = 10 * time.Second
	maxAttempts  = 8
)

// RetryOpen opens a Manager with exponential backoff. Returns ErrConnectionDeadline
// after maxAttempts or when ctx is cancelled.
func RetryOpen(ctx context.Context, dsn string, log *slog.Logger) (manager.Manager, error) {
	delay := initialDelay
	for attempt := 1; ; attempt++ {
		mgr, err := localdatabase.OpenManager(ctx, dsn)
		if err == nil {
			return mgr, nil
		}
		if attempt >= maxAttempts || ctx.Err() != nil {
			return nil, fmt.Errorf("%w after %d attempts: %w", ErrConnectionDeadline, attempt, err)
		}
		log.Warn("connection failed, will retry",
			slog.Int("attempt", attempt),
			slog.String("delay", delay.String()),
			slog.String("error", err.Error()),
		)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", ErrConnectionDeadline, ctx.Err())
		case <-time.After(delay):
		}
		if delay < maxDelay {
			delay *= 2
		}
	}
}
