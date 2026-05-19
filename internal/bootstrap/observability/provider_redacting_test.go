package observability

import (
	"context"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

func TestWrapWithRedaction_LoggerRedactsSensitiveFields(t *testing.T) {
	base := fake.NewProvider()
	obs := wrapWithRedaction(base, redactor.DefaultDenylist)

	obs.Logger().Info(context.Background(), "test",
		observability.String("cpf", "12345678900"),
		observability.Any("payload", map[string]any{
			"email": "user@example.com",
			"safe":  "ok",
		}),
	)

	entries := base.Logger().(*fake.FakeLogger).GetEntries()
	require.Len(t, entries, 1)

	fields := entries[0].Fields
	assert.Equal(t, redactor.Sentinel, fieldValue(t, fields, "cpf"))

	payload, ok := fieldAny(t, fields, "payload").(map[string]any)
	require.True(t, ok)
	assert.Equal(t, redactor.Sentinel, payload["email"])
	assert.Equal(t, "ok", payload["safe"])
}

func TestWrapWithRedaction_TracerRedactsSpanAttributesAndEvents(t *testing.T) {
	base := fake.NewProvider()
	obs := wrapWithRedaction(base, redactor.DefaultDenylist)

	ctx, span := obs.Tracer().Start(context.Background(), "test-span",
		observability.WithAttributes(observability.String("token", "secret-token")),
	)
	span.SetAttributes(observability.String("authorization", "Bearer abc"))
	span.AddEvent("evt", observability.String("email", "user@example.com"))
	obs.Tracer().SpanFromContext(ctx).SetAttributes(observability.String("cpf", "12345678900"))
	span.End()

	spans := base.Tracer().(*fake.FakeTracer).GetSpans()
	require.Len(t, spans, 1)

	assert.Equal(t, redactor.Sentinel, fieldValue(t, spans[0].Attributes, "token"))
	assert.Equal(t, redactor.Sentinel, fieldValue(t, spans[0].Attributes, "authorization"))
	assert.Equal(t, redactor.Sentinel, fieldValue(t, spans[0].Attributes, "cpf"))
	require.Len(t, spans[0].Events, 1)
	assert.Equal(t, redactor.Sentinel, fieldValue(t, spans[0].Events[0].Fields, "email"))
}

func fieldValue(t *testing.T, fields []observability.Field, key string) string {
	t.Helper()
	for _, field := range fields {
		if field.Key == key {
			return field.StringValue()
		}
	}
	t.Fatalf("field %q not found", key)
	return ""
}

func fieldAny(t *testing.T, fields []observability.Field, key string) any {
	t.Helper()
	for _, field := range fields {
		if field.Key == key {
			return field.AnyValue()
		}
	}
	t.Fatalf("field %q not found", key)
	return nil
}
