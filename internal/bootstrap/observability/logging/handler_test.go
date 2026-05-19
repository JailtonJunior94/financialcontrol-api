package logging_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/logging"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

func newTestHandler(buf *bytes.Buffer) *logging.Handler {
	downstream := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return logging.New(redactor.DefaultDenylist, downstream)
}

func TestHandler_PII_FieldRedacted(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantPII bool
	}{
		{
			name:    "cpf field is redacted",
			key:     "cpf",
			value:   "123.456.789-00",
			wantPII: false,
		},
		{
			name:    "email field is redacted",
			key:     "email",
			value:   "user@example.com",
			wantPII: false,
		},
		{
			name:    "password field is redacted",
			key:     "password",
			value:   "s3cr3t",
			wantPII: false,
		},
		{
			name:    "token field is redacted",
			key:     "token",
			value:   "eyJhbGciOiJIUzI1NiJ9",
			wantPII: false,
		},
		{
			name:    "safe field is preserved",
			key:     "user_id",
			value:   "user-42",
			wantPII: true,
		},
		{
			name:    "transaction_id is preserved",
			key:     "transaction_id",
			value:   "txn-123",
			wantPII: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			h := newTestHandler(&buf)
			logger := slog.New(h)

			logger.Info("test message", tc.key, tc.value)

			out := buf.String()
			if tc.wantPII {
				assert.Contains(t, out, tc.value)
				assert.NotContains(t, out, redactor.Sentinel)
			} else {
				assert.NotContains(t, out, tc.value)
				assert.Contains(t, out, redactor.Sentinel)
			}
		})
	}
}

func TestHandler_Enabled_DelegatesToDownstream(t *testing.T) {
	var buf bytes.Buffer
	downstream := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := logging.New(redactor.DefaultDenylist, downstream)

	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
	assert.False(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, h.Enabled(context.Background(), slog.LevelError))
}

func TestHandler_MessagePreserved(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	logger := slog.New(h)

	logger.Info("operation completed", "cpf", "123.456.789-00", "amount", 100)

	out := buf.String()
	assert.Contains(t, out, "operation completed")
	assert.Contains(t, out, "100")
	assert.NotContains(t, out, "123.456.789-00")
	assert.Contains(t, out, redactor.Sentinel)
}

func TestHandler_WithAttrs_RedactsPIIAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	enriched := h.WithAttrs([]slog.Attr{
		slog.String("email", "user@example.com"),
		slog.String("request_id", "req-1"),
	})
	logger := slog.New(enriched)

	logger.Info("test")

	out := buf.String()
	assert.NotContains(t, out, "user@example.com")
	assert.Contains(t, out, redactor.Sentinel)
	assert.Contains(t, out, "req-1")
}

func TestHandler_WithGroup_PreservesGroup(t *testing.T) {
	var buf bytes.Buffer
	h := newTestHandler(&buf)
	grouped := h.WithGroup("http")
	logger := slog.New(grouped)

	logger.Info("request", "cpf", "999")

	out := buf.String()
	assert.Contains(t, out, "http")
	assert.NotContains(t, out, "999")
}

func TestHandler_Handle_TimestampPreserved(t *testing.T) {
	var buf bytes.Buffer
	downstream := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	h := logging.New(redactor.DefaultDenylist, downstream)

	ts := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	record := slog.NewRecord(ts, slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("token", "abc"))

	require.NoError(t, h.Handle(context.Background(), record))

	out := buf.String()
	assert.Contains(t, out, "2026-05-14")
	assert.NotContains(t, out, "abc")
	assert.Contains(t, out, redactor.Sentinel)
}
