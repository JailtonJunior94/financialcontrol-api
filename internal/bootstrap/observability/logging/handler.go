package logging

import (
	"context"
	"log/slog"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

// Handler is a slog.Handler that redacts PII attributes before delegating to downstream.
type Handler struct {
	downstream slog.Handler
	denylist   *redactor.Denylist
}

// New returns a Handler wrapping downstream. Each slog.Record is cloned with PII
// attributes replaced by redactor.Sentinel before being forwarded.
func New(dl *redactor.Denylist, downstream slog.Handler) *Handler {
	return &Handler{downstream: downstream, denylist: dl}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.downstream.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	redacted := redactor.RedactAttrs(h.denylist, attrs,
		func(a slog.Attr) string { return a.Key },
		func(a slog.Attr) slog.Attr { return slog.String(a.Key, redactor.Sentinel) },
	)
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	clean.AddAttrs(redacted...)
	return h.downstream.Handle(ctx, clean)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := redactor.RedactAttrs(h.denylist, attrs,
		func(a slog.Attr) string { return a.Key },
		func(a slog.Attr) slog.Attr { return slog.String(a.Key, redactor.Sentinel) },
	)
	return &Handler{
		downstream: h.downstream.WithAttrs(redacted),
		denylist:   h.denylist,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		downstream: h.downstream.WithGroup(name),
		denylist:   h.denylist,
	}
}
