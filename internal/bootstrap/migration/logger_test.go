package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactingHandler_SensitiveKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		key      string
		value    string
		redacted bool
	}{
		{"password key", "password", "secret123", true},
		{"dsn key", "dsn", "sqlserver://sa:pass@host", true},
		{"connection_string key", "connection_string", "sqlserver://sa:pass@host", true},
		{"token key", "token", "abc123", true},
		{"secret key", "secret", "mysecret", true},
		{"safe key database", "database", "mydb", false},
		{"safe key host", "host", "localhost", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			handler := &redactingHandler{
				inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
			}
			log := slog.New(handler)
			log.Info("test", slog.String(tc.key, tc.value))

			var record map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &record))

			if tc.redacted {
				assert.Equal(t, "***", record[tc.key], "expected redaction for key %q", tc.key)
			} else {
				assert.Equal(t, tc.value, record[tc.key], "expected value for key %q", tc.key)
			}
		})
	}
}

func TestRedactingHandler_JSONFormat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	handler := &redactingHandler{
		inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}
	log := slog.New(handler)
	log.Info("test message", slog.String("key", "value"))

	var record map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	assert.Equal(t, "test message", record["msg"])
	assert.Equal(t, "value", record["key"])
}

func TestRedactingHandler_AuditFieldsPropagated(t *testing.T) {
	t.Setenv("CI_BUILD_ID", "build-42")

	var buf bytes.Buffer
	handler := &redactingHandler{
		inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}
	log := slog.New(handler)
	for _, attr := range AuditFields() {
		log = log.With(attr)
	}
	log.Info("audit test")

	var record map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	assert.Equal(t, "build-42", record["build_id"])
}

func TestRedactingHandler_WithAttrs_RedactsAndPasses(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := &redactingHandler{
		inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}
	log := slog.New(h.WithAttrs([]slog.Attr{
		slog.String("password", "secret"),
		slog.String("host", "localhost"),
	}))
	log.Info("test")

	var record map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	assert.Equal(t, "***", record["password"])
	assert.Equal(t, "localhost", record["host"])
}

func TestRedactingHandler_Enabled(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	handler := &redactingHandler{
		inner: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}),
	}
	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, handler.Enabled(context.Background(), slog.LevelDebug))
}
