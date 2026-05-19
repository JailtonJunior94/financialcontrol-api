package observability

import (
	"context"

	devkitobs "github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

type redactingProvider struct {
	base   devkitobs.Observability
	logger devkitobs.Logger
	tracer devkitobs.Tracer
}

func wrapWithRedaction(base devkitobs.Observability, dl *redactor.Denylist) devkitobs.Observability {
	if base == nil || dl == nil {
		return base
	}
	return &redactingProvider{
		base:   base,
		logger: &redactingLogger{base: base.Logger(), dl: dl},
		tracer: &redactingTracer{base: base.Tracer(), dl: dl},
	}
}

func (p *redactingProvider) Tracer() devkitobs.Tracer   { return p.tracer }
func (p *redactingProvider) Logger() devkitobs.Logger   { return p.logger }
func (p *redactingProvider) Metrics() devkitobs.Metrics { return p.base.Metrics() }
func (p *redactingProvider) Shutdown(ctx context.Context) error {
	return p.base.Shutdown(ctx)
}

type redactingLogger struct {
	base devkitobs.Logger
	dl   *redactor.Denylist
}

func (l *redactingLogger) Debug(ctx context.Context, msg string, fields ...devkitobs.Field) {
	l.base.Debug(ctx, msg, redactFields(l.dl, fields)...)
}

func (l *redactingLogger) Info(ctx context.Context, msg string, fields ...devkitobs.Field) {
	l.base.Info(ctx, msg, redactFields(l.dl, fields)...)
}

func (l *redactingLogger) Warn(ctx context.Context, msg string, fields ...devkitobs.Field) {
	l.base.Warn(ctx, msg, redactFields(l.dl, fields)...)
}

func (l *redactingLogger) Error(ctx context.Context, msg string, fields ...devkitobs.Field) {
	l.base.Error(ctx, msg, redactFields(l.dl, fields)...)
}

func (l *redactingLogger) With(fields ...devkitobs.Field) devkitobs.Logger {
	return &redactingLogger{
		base: l.base.With(redactFields(l.dl, fields)...),
		dl:   l.dl,
	}
}

type redactingTracer struct {
	base devkitobs.Tracer
	dl   *redactor.Denylist
}

func (t *redactingTracer) Start(ctx context.Context, spanName string, opts ...devkitobs.SpanOption) (context.Context, devkitobs.Span) {
	cfg := devkitobs.NewSpanConfig(opts)
	redactedOpts := []devkitobs.SpanOption{devkitobs.WithSpanKind(cfg.Kind())}
	if attrs := redactFields(t.dl, cfg.Attributes()); len(attrs) > 0 {
		redactedOpts = append(redactedOpts, devkitobs.WithAttributes(attrs...))
	}

	ctx, span := t.base.Start(ctx, spanName, redactedOpts...)
	wrapped := &redactingSpan{base: span, dl: t.dl}
	return ctx, wrapped
}

func (t *redactingTracer) SpanFromContext(ctx context.Context) devkitobs.Span {
	return &redactingSpan{base: t.base.SpanFromContext(ctx), dl: t.dl}
}

func (t *redactingTracer) ContextWithSpan(ctx context.Context, span devkitobs.Span) context.Context {
	if wrapped, ok := span.(*redactingSpan); ok {
		return t.base.ContextWithSpan(ctx, wrapped.base)
	}
	return t.base.ContextWithSpan(ctx, span)
}

type redactingSpan struct {
	base devkitobs.Span
	dl   *redactor.Denylist
}

func (s *redactingSpan) End() { s.base.End() }

func (s *redactingSpan) SetAttributes(fields ...devkitobs.Field) {
	s.base.SetAttributes(redactFields(s.dl, fields)...)
}

func (s *redactingSpan) SetStatus(code devkitobs.StatusCode, description string) {
	s.base.SetStatus(code, description)
}

func (s *redactingSpan) RecordError(err error, fields ...devkitobs.Field) {
	s.base.RecordError(err, redactFields(s.dl, fields)...)
}

func (s *redactingSpan) AddEvent(name string, fields ...devkitobs.Field) {
	s.base.AddEvent(name, redactFields(s.dl, fields)...)
}

func (s *redactingSpan) Context() devkitobs.SpanContext { return s.base.Context() }
func (s *redactingSpan) TraceID() string                { return s.base.TraceID() }
func (s *redactingSpan) SpanID() string                 { return s.base.SpanID() }
func (s *redactingSpan) IsSampled() bool                { return s.base.IsSampled() }

func redactFields(dl *redactor.Denylist, fields []devkitobs.Field) []devkitobs.Field {
	if len(fields) == 0 {
		return nil
	}

	out := make([]devkitobs.Field, len(fields))
	for i, field := range fields {
		if dl.Contains(field.Key) {
			out[i] = devkitobs.String(field.Key, redactor.Sentinel)
			continue
		}

		if field.Kind() == devkitobs.FieldKindAny {
			out[i] = devkitobs.Any(field.Key, redactAny(dl, field.AnyValue()))
			continue
		}

		out[i] = field
	}
	return out
}

func redactAny(dl *redactor.Denylist, value any) any {
	switch v := value.(type) {
	case map[string]any:
		return dl.RedactMap(v)
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = redactAny(dl, v[i])
		}
		return out
	default:
		return value
	}
}
