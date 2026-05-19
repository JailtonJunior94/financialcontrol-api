package usecase

import (
	"context"
	"log/slog"
)

// safeRecord calls fn and recovers from any panic, emitting a WARN log.
// Recorder methods return void, so panics are the only failure mode to handle.
func safeRecord(ctx context.Context, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			slog.WarnContext(ctx, "metrics recorder panicked", "recover", r)
		}
	}()
	fn()
}
