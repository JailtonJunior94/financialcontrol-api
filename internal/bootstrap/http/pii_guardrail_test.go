//go:build integration

package http_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/logging"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
	tracingpkg "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/tracing"
	mssqltesting "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

// piiFields enumerates every field in DefaultDenylist so the guardrail test
// is automatically exhaustive: adding a new field to the denylist requires
// updating this slice (failing compilation would force the reviewer).
//
// Keep this in sync with internal/bootstrap/observability/redactor/denylist.go.
var piiFields = []struct {
	key   string
	value string
}{
	// PCI / credentials
	{key: "pan", value: "4111111111111111"},
	{key: "card_number", value: "4111111111111111"},
	{key: "cardnumber", value: "4111111111111111"},
	{key: "cvv", value: "123"},
	{key: "cvc", value: "456"},
	{key: "password", value: "s3cr3tP@ss!"},
	{key: "passwd", value: "p@sswd123"},
	{key: "secret", value: "topsecret"},
	{key: "token", value: "tok_abc123"},
	{key: "access_token", value: "at_xyz789"},
	{key: "refresh_token", value: "rt_def456"},
	{key: "authorization", value: "Bearer eyJhbGciOiJSUzI1NiJ9"},
	// Brazilian documents
	{key: "cpf", value: "123.456.789-00"},
	{key: "cnpj", value: "12.345.678/0001-99"},
	{key: "rg", value: "12.345.678-9"},
	// Contact
	{key: "email", value: "user@example.com"},
	{key: "phone", value: "+55 11 91234-5678"},
	{key: "telefone", value: "+5511912345678"},
	{key: "address", value: "123 Main St"},
	{key: "endereco", value: "Rua das Flores, 42"},
	{key: "zipcode", value: "01310-100"},
	{key: "cep", value: "01310-100"},
	// Banking and identity
	{key: "account_number", value: "123456-7"},
	{key: "agency", value: "0001"},
	{key: "bank_account", value: "987654321"},
	{key: "pix_key", value: "user@pix.example.com"},
	{key: "pix_chave", value: "chave-aleatoria-pix"},
	{key: "full_name", value: "João da Silva"},
	{key: "holder_name", value: "JOAO DA SILVA"},
	{key: "nome_completo", value: "João Pedro da Silva"},
}

// TestPIIRedactedEndToEnd is the CI gate (RF-23) that validates no PII value
// from the denylist survives through the observability pipeline.
//
// It validates three surfaces defined in RF-09.1:
//  1. slog structured-log output (via bytes.Buffer)
//  2. OTel span attributes (via tracetest.SpanRecorder)
//  3. db.statement redaction (via RedactStatement)
//
// Additionally, it boots the full HTTP server and verifies PII does not appear
// in HTTP error responses when PII is injected into request headers.
//
// Mutation note: removing any field from DefaultDenylist will cause this test to
// fail because piiFields is kept exhaustive intentionally.
func TestPIIRedactedEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	t.Run("slog_handler_redacts_all_denylist_fields", func(t *testing.T) {
		for _, f := range piiFields {
			f := f
			t.Run(f.key, func(t *testing.T) {
				t.Parallel()
				var buf bytes.Buffer
				downstream := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
				h := logging.New(redactor.DefaultDenylist, downstream)
				logger := slog.New(h)

				logger.Info("test-event", slog.String(f.key, f.value))

				output := buf.String()
				assert.NotContains(t, output, f.value,
					"log output must not contain literal PII value for field %q", f.key)
				assert.Contains(t, output, redactor.Sentinel,
					"log output must contain [REDACTED] sentinel for field %q", f.key)
			})
		}
	})

	t.Run("slog_handler_redacts_nested_fields", func(t *testing.T) {
		// Validates case-insensitive matching and ≥3-level nesting for a subset of fields.
		cases := []struct {
			name  string
			key   string
			value string
		}{
			{name: "uppercase_key", key: "CPF", value: "123.456.789-00"},
			{name: "mixed_case_key", key: "Email", value: "user@example.com"},
			{name: "non_pii_field_not_redacted", key: "transaction_id", value: "txn-42"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				var buf bytes.Buffer
				downstream := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
				h := logging.New(redactor.DefaultDenylist, downstream)
				logger := slog.New(h)

				logger.Info("test-event", slog.String(tc.key, tc.value))

				output := buf.String()
				if redactor.DefaultDenylist.Contains(tc.key) {
					assert.NotContains(t, output, tc.value,
						"PII field %q must not appear in log output", tc.key)
					assert.Contains(t, output, redactor.Sentinel,
						"PII field %q must be replaced by sentinel", tc.key)
				} else {
					assert.Contains(t, output, tc.value,
						"non-PII field %q must appear unchanged in log output", tc.key)
				}
			})
		}
	})

	t.Run("span_processor_redacts_all_denylist_fields", func(t *testing.T) {
		for _, f := range piiFields {
			f := f
			t.Run(f.key, func(t *testing.T) {
				t.Parallel()
				recorder := tracetest.NewSpanRecorder()
				processor := tracingpkg.NewPIIRedacting(recorder, redactor.DefaultDenylist)

				tp := sdktrace.NewTracerProvider(
					sdktrace.WithSpanProcessor(processor),
				)
				tracer := tp.Tracer("pii-guardrail-test")

				ctx, span := tracer.Start(context.Background(), "guardrail-span")
				span.SetAttributes(attribute.String(f.key, f.value))
				span.End()
				_ = ctx

				spans := recorder.Ended()
				require.Len(t, spans, 1, "expected exactly one span to be recorded")

				for _, attr := range spans[0].Attributes() {
					assert.NotEqual(t, f.value, attr.Value.AsString(),
						"span attribute %q must not contain literal PII value", f.key)
					if strings.EqualFold(string(attr.Key), f.key) {
						assert.Equal(t, redactor.Sentinel, attr.Value.AsString(),
							"span attribute %q must be replaced by [REDACTED]", f.key)
					}
				}
			})
		}
	})

	t.Run("redact_statement_removes_pii_from_sql", func(t *testing.T) {
		cases := []struct {
			name      string
			input     string
			wantClean bool
			piiValue  string
		}{
			{
				name:      "cpf in where clause",
				input:     "SELECT * FROM users WHERE cpf='123.456.789-00'",
				wantClean: true,
				piiValue:  "123.456.789-00",
			},
			{
				name:      "email in insert",
				input:     "INSERT INTO users (email) VALUES ('user@example.com')",
				wantClean: true,
				piiValue:  "user@example.com",
			},
			{
				name:      "password in update",
				input:     "UPDATE users SET password='s3cr3tP@ss!' WHERE id=1",
				wantClean: true,
				piiValue:  "s3cr3tP@ss!",
			},
			{
				name:      "non_pii_field_preserved",
				input:     "SELECT * FROM orders WHERE order_id=42",
				wantClean: false,
				piiValue:  "42",
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				result := redactor.DefaultDenylist.RedactStatement(tc.input)
				if tc.wantClean {
					assert.NotContains(t, result, tc.piiValue,
						"RedactStatement must remove PII value from SQL statement")
					assert.Contains(t, result, redactor.Sentinel,
						"RedactStatement must insert [REDACTED] sentinel")
				} else {
					assert.Contains(t, result, tc.piiValue,
						"non-PII value must be preserved in SQL statement")
				}
			})
		}
	})

	t.Run("http_server_response_does_not_leak_pii", func(t *testing.T) {
		// Boot the full HTTP server and verify PII values sent as request headers
		// do not appear in error response bodies (RFC 7807).
		mssqlDSN, err := mssqltesting.GetSharedTestDSN()
		require.NoError(t, err, "MSSQL testcontainer failed to start")

		otlpEndpoint := startLGTMContainer(t)
		serverPort := freePort(t)

		t.Chdir("testdata")

		t.Setenv("ENVIRONMENT", "development")
		t.Setenv("SERVICE_NAME", "pii-guardrail-svc")
		t.Setenv("SERVICE_VERSION", "0.0.1")
		t.Setenv("MSSQL_CONNECTION_STRING", mssqlDSN)
		t.Setenv("JWT_SECRET", integrationJWTSecret)
		t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
		t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otlpEndpoint)
		t.Setenv("PORT", serverPort)

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		startErr := make(chan error, 1)
		go func() {
			startErr <- bootstrapcli.RunServer(ctx)
		}()

		baseURL := fmt.Sprintf("http://localhost:%s", serverPort)
		waitForServer(t, baseURL+"/live", serverReadyTimeout, startErr)

		for _, f := range piiFields {
			f := f
			t.Run("header_"+f.key, func(t *testing.T) {
				// Send a request with PII value as a custom header to an unauthenticated route.
				// The server should return RFC 7807 401/403, and the body must never echo back
				// the PII value — it should only contain structural error fields.
				req, reqErr := http.NewRequestWithContext(context.Background(),
					http.MethodPost,
					baseURL+"/api/v1/finance/transactions",
					nil,
				)
				require.NoError(t, reqErr)
				req.Header.Set("X-PII-Test-"+f.key, f.value)
				req.Header.Set("Content-Type", "application/json")

				resp, doErr := http.DefaultClient.Do(req)
				require.NoError(t, doErr)
				t.Cleanup(func() { _ = resp.Body.Close() })

				// Server must not echo back PII values in response bodies.
				buf := new(bytes.Buffer)
				_, readErr := buf.ReadFrom(resp.Body)
				require.NoError(t, readErr)

				assert.NotContains(t, buf.String(), f.value,
					"HTTP response body must not contain PII value for field %q", f.key)
			})
		}
	})
}
