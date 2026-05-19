package tracing

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

// PIIRedactingSpanProcessor wraps a downstream SpanProcessor. In OnEnd it builds a
// redactedSpan view that replaces denylist-matching attribute values with [REDACTED]
// before forwarding to the downstream processor.
type PIIRedactingSpanProcessor struct {
	downstream sdktrace.SpanProcessor
	denylist   *redactor.Denylist
}

// NewPIIRedacting returns a PIIRedactingSpanProcessor wrapping downstream.
func NewPIIRedacting(downstream sdktrace.SpanProcessor, dl *redactor.Denylist) *PIIRedactingSpanProcessor {
	return &PIIRedactingSpanProcessor{downstream: downstream, denylist: dl}
}

func (p *PIIRedactingSpanProcessor) OnStart(ctx context.Context, s sdktrace.ReadWriteSpan) {
	p.downstream.OnStart(ctx, s)
}

func (p *PIIRedactingSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	p.downstream.OnEnd(&redactedSpan{ReadOnlySpan: s, dl: p.denylist})
}

func (p *PIIRedactingSpanProcessor) Shutdown(ctx context.Context) error {
	return p.downstream.Shutdown(ctx)
}

func (p *PIIRedactingSpanProcessor) ForceFlush(ctx context.Context) error {
	return p.downstream.ForceFlush(ctx)
}

// redactedSpan wraps a ReadOnlySpan and returns PII-redacted Attributes and Events.
// Embedding the interface satisfies the unexported private() method via delegation.
type redactedSpan struct {
	sdktrace.ReadOnlySpan
	dl *redactor.Denylist
}

func (r *redactedSpan) Attributes() []attribute.KeyValue {
	return redactor.RedactAttrs(r.dl, r.ReadOnlySpan.Attributes(),
		func(kv attribute.KeyValue) string { return string(kv.Key) },
		func(kv attribute.KeyValue) attribute.KeyValue {
			return attribute.String(string(kv.Key), redactor.Sentinel)
		},
	)
}

func (r *redactedSpan) Events() []sdktrace.Event {
	orig := r.ReadOnlySpan.Events()
	out := make([]sdktrace.Event, len(orig))
	for i, ev := range orig {
		redactedAttrs := redactor.RedactAttrs(r.dl, ev.Attributes,
			func(kv attribute.KeyValue) string { return string(kv.Key) },
			func(kv attribute.KeyValue) attribute.KeyValue {
				return attribute.String(string(kv.Key), redactor.Sentinel)
			},
		)
		out[i] = sdktrace.Event{
			Name:                  ev.Name,
			Attributes:            redactedAttrs,
			DroppedAttributeCount: ev.DroppedAttributeCount,
			Time:                  ev.Time,
		}
	}
	return out
}

const warnThrottle = 60 * time.Second

// ThrottledSpanExporter wraps a SpanExporter and throttles WARN logs on export failure.
// At most one WARN per warnThrottle window. ExportSpans always returns nil (RF-15.6).
type ThrottledSpanExporter struct {
	downstream        sdktrace.SpanExporter
	logger            *slog.Logger
	mu                sync.Mutex
	lastWarn          time.Time
	dropped           atomic.Int64
	droppedAtLastWarn atomic.Int64
}

// NewThrottledExporter returns a ThrottledSpanExporter wrapping downstream.
func NewThrottledExporter(downstream sdktrace.SpanExporter, logger *slog.Logger) *ThrottledSpanExporter {
	return &ThrottledSpanExporter{downstream: downstream, logger: logger}
}

// ExportSpans forwards to downstream. On error, increments the dropped counter and emits
// at most one WARN per 60s window. Always returns nil so callers are never blocked.
func (t *ThrottledSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if err := t.downstream.ExportSpans(ctx, spans); err != nil {
		total := t.dropped.Add(1)

		t.mu.Lock()
		shouldWarn := time.Since(t.lastWarn) >= warnThrottle
		if shouldWarn {
			t.lastWarn = time.Now()
		}
		t.mu.Unlock()

		if shouldWarn {
			prev := t.droppedAtLastWarn.Swap(total)
			sinceLast := total - prev
			t.logger.Warn("otlp_export_failed",
				slog.String("event", "otlp_export_failed"),
				slog.Int64("dropped_since_last_warn", sinceLast),
				slog.String("error", err.Error()),
			)
		}
	}
	return nil
}

func (t *ThrottledSpanExporter) Shutdown(ctx context.Context) error {
	return t.downstream.Shutdown(ctx)
}

// Dropped returns the total number of failed export calls since this exporter was created.
func (t *ThrottledSpanExporter) Dropped() int64 {
	return t.dropped.Load()
}

// noopSpanProcessor is a SpanProcessor that discards all spans.
type noopSpanProcessor struct{}

func (noopSpanProcessor) OnStart(_ context.Context, _ sdktrace.ReadWriteSpan) {}
func (noopSpanProcessor) OnEnd(_ sdktrace.ReadOnlySpan)                       {}
func (noopSpanProcessor) Shutdown(_ context.Context) error                    { return nil }
func (noopSpanProcessor) ForceFlush(_ context.Context) error                  { return nil }

// NoopProcessor returns a SpanProcessor that discards all spans.
func NoopProcessor() sdktrace.SpanProcessor {
	return noopSpanProcessor{}
}
