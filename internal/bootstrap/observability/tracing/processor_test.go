package tracing_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/tracing"
)

// failExporter always returns an error on ExportSpans.
type failExporter struct{}

func (f *failExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return errors.New("simulated export failure")
}

func (f *failExporter) Shutdown(_ context.Context) error { return nil }

// spanFromStub creates a ReadOnlySpan from a SpanStub.
func spanFromStub(stub tracetest.SpanStub) sdktrace.ReadOnlySpan {
	return stub.Snapshot()
}

// TestPIIRedactingSpanProcessor_AttributesRedacted verifies that PII attribute values are
// replaced by redactor.Sentinel before being forwarded to the downstream processor.
func TestPIIRedactingSpanProcessor_AttributesRedacted(t *testing.T) {
	tests := []struct {
		name        string
		inputAttrs  []attribute.KeyValue
		redactedKey string
		safeKey     string
		safeValue   string
	}{
		{
			name: "cpf attribute is redacted",
			inputAttrs: []attribute.KeyValue{
				attribute.String("cpf", "123.456.789-00"),
				attribute.String("user_id", "user-42"),
			},
			redactedKey: "cpf",
			safeKey:     "user_id",
			safeValue:   "user-42",
		},
		{
			name: "token attribute is redacted",
			inputAttrs: []attribute.KeyValue{
				attribute.String("token", "eyJhbGciOiJIUzI1NiJ9"),
				attribute.String("operation", "create_transaction"),
			},
			redactedKey: "token",
			safeKey:     "operation",
			safeValue:   "create_transaction",
		},
		{
			name: "email attribute is redacted",
			inputAttrs: []attribute.KeyValue{
				attribute.String("email", "pii@example.com"),
				attribute.String("amount", "100.00"),
			},
			redactedKey: "email",
			safeKey:     "amount",
			safeValue:   "100.00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			p := tracing.NewPIIRedacting(recorder, redactor.DefaultDenylist)

			stub := tracetest.SpanStub{Name: "test-span", Attributes: tc.inputAttrs}
			p.OnEnd(spanFromStub(stub))

			ended := recorder.Ended()
			require.Len(t, ended, 1)

			attrMap := attrsToMap(ended[0].Attributes())
			assert.Equal(t, redactor.Sentinel, attrMap[tc.redactedKey], "PII attribute must be redacted")
			assert.Equal(t, tc.safeValue, attrMap[tc.safeKey], "non-PII attribute must be preserved")
		})
	}
}

// TestPIIRedactingSpanProcessor_EventsRedacted verifies that PII inside span events is redacted.
func TestPIIRedactingSpanProcessor_EventsRedacted(t *testing.T) {
	stub := tracetest.SpanStub{
		Name: "event-span",
		Events: []sdktrace.Event{
			{
				Name: "user.action",
				Attributes: []attribute.KeyValue{
					attribute.String("email", "pii@example.com"),
					attribute.String("action", "login"),
				},
			},
		},
	}

	recorder := tracetest.NewSpanRecorder()
	p := tracing.NewPIIRedacting(recorder, redactor.DefaultDenylist)
	p.OnEnd(spanFromStub(stub))

	ended := recorder.Ended()
	require.Len(t, ended, 1)
	events := ended[0].Events()
	require.Len(t, events, 1)

	eventAttrMap := attrsToMap(events[0].Attributes)
	assert.Equal(t, redactor.Sentinel, eventAttrMap["email"], "email in event must be redacted")
	assert.Equal(t, "login", eventAttrMap["action"], "action in event must be preserved")
}

// TestPIIRedactingSpanProcessor_NoAttributeSpan does not panic on a span without attributes.
func TestPIIRedactingSpanProcessor_NoAttributeSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	p := tracing.NewPIIRedacting(recorder, redactor.DefaultDenylist)

	assert.NotPanics(t, func() {
		p.OnEnd(spanFromStub(tracetest.SpanStub{Name: "empty"}))
	})
	assert.Len(t, recorder.Ended(), 1)
}

// TestPIIRedactingSpanProcessor_DelegatesLifecycle verifies Shutdown/ForceFlush are forwarded.
func TestPIIRedactingSpanProcessor_DelegatesLifecycle(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	p := tracing.NewPIIRedacting(recorder, redactor.DefaultDenylist)

	require.NoError(t, p.Shutdown(context.Background()))
	require.NoError(t, p.ForceFlush(context.Background()))
}

// TestThrottledExporter_100Failures_ExactlyOneWarn verifies that 100 rapid consecutive
// export failures produce exactly one WARN and that Dropped() returns 100.
func TestThrottledExporter_100Failures_ExactlyOneWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	te := tracing.NewThrottledExporter(&failExporter{}, logger)

	ctx := context.Background()
	for range 100 {
		require.NoError(t, te.ExportSpans(ctx, nil))
	}

	assert.Equal(t, int64(100), te.Dropped(), "total dropped counter must equal 100")

	warnCount := countLines(buf.String(), "otlp_export_failed")
	assert.Equal(t, 1, warnCount, "exactly one WARN must be emitted within a 60s window")
}

// TestThrottledExporter_DoesNotBlockCaller verifies RF-15.6: ExportSpans always returns nil.
func TestThrottledExporter_DoesNotBlockCaller(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	te := tracing.NewThrottledExporter(&failExporter{}, logger)

	start := time.Now()
	err := te.ExportSpans(context.Background(), nil)
	elapsed := time.Since(start)

	require.NoError(t, err, "export failure must not propagate to caller")
	assert.Less(t, elapsed, 100*time.Millisecond, "must not block caller")
}

// TestThrottledExporter_NoFailureNoWarn verifies no WARN on successful export.
func TestThrottledExporter_NoFailureNoWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	te := tracing.NewThrottledExporter(tracetest.NewNoopExporter(), logger)

	for range 10 {
		require.NoError(t, te.ExportSpans(context.Background(), nil))
	}

	assert.Empty(t, buf.String(), "no WARN should be emitted when export succeeds")
	assert.Equal(t, int64(0), te.Dropped())
}

func attrsToMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}

// TestPIIRedactingSpanProcessor_OnStart_Forwarded verifies that OnStart is forwarded to the downstream.
func TestPIIRedactingSpanProcessor_OnStart_Forwarded(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	p := tracing.NewPIIRedacting(recorder, redactor.DefaultDenylist)

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "started-span")
	span.End()

	// Verify that the span ended up in the recorder (proving OnStart + OnEnd chain works).
	assert.Len(t, recorder.Ended(), 1)
}

// TestThrottledExporter_Shutdown_Forwarded verifies that Shutdown is delegated to downstream.
func TestThrottledExporter_Shutdown_Forwarded(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	te := tracing.NewThrottledExporter(tracetest.NewNoopExporter(), logger)

	require.NoError(t, te.Shutdown(context.Background()))
}

// TestNoopProcessor_Satisfies_SpanProcessor verifies that NoopProcessor() returns a
// valid SpanProcessor that can be used as a downstream without panics.
func TestNoopProcessor_Satisfies_SpanProcessor(t *testing.T) {
	p := tracing.NewPIIRedacting(tracing.NoopProcessor(), redactor.DefaultDenylist)

	stub := tracetest.SpanStub{
		Name: "noop-test",
		Attributes: []attribute.KeyValue{
			attribute.String("cpf", "999"),
		},
	}
	assert.NotPanics(t, func() {
		p.OnEnd(stub.Snapshot())
	})
}

func countLines(s, substr string) int {
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if strings.Contains(line, substr) {
			count++
		}
	}
	return count
}
