package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/metrics"
	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/middleware"
)

func newTestBusinessMetrics(t *testing.T) (*metrics.BusinessMetrics, *fake.FakeMetrics) {
	t.Helper()
	fp := fake.NewProvider()
	bm, err := metrics.NewBusinessMetrics(fp)
	require.NoError(t, err)
	fm := fp.Metrics().(*fake.FakeMetrics)
	return bm, fm
}

// buildApp creates a Fiber test app with FinanceMetrics middleware at /finance prefix
// and registers a single route with the given method, path, status and optional name.
func buildApp(bm *metrics.BusinessMetrics, method, path string, status int, name string) *fiber.App {
	app := fiber.New()
	app.Use("/finance", middleware.FinanceMetrics(bm))

	handler := func(c *fiber.Ctx) error {
		return c.SendStatus(status)
	}

	route := app.Add(method, path, handler)
	if name != "" {
		route.Name(name)
	}
	return app
}

// buildAppWithContext creates a Fiber test app that injects the provided context
// into the Fiber request context before calling the middleware chain.
func buildAppWithContext(bm *metrics.BusinessMetrics, method, path string, status int, name string, ctx context.Context) *fiber.App {
	app := fiber.New()

	// Inject the caller-provided context so spans added via trace.SpanFromContext work.
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(ctx)
		return c.Next()
	})

	app.Use("/finance", middleware.FinanceMetrics(bm))

	handler := func(c *fiber.Ctx) error {
		return c.SendStatus(status)
	}

	route := app.Add(method, path, handler)
	if name != "" {
		route.Name(name)
	}
	return app
}

func buildAppReturningError(bm *metrics.BusinessMetrics, method, path string, err error, name string) *fiber.App {
	app := fiber.New()
	app.Use("/finance", middleware.FinanceMetrics(bm))

	handler := func(c *fiber.Ctx) error {
		return err
	}

	route := app.Add(method, path, handler)
	if name != "" {
		route.Name(name)
	}
	return app
}

func hasStringField(fields []observability.Field, key, val string) bool {
	for _, f := range fields {
		if f.Key == key && f.Kind() == observability.FieldKindString && f.StringValue() == val {
			return true
		}
	}
	return false
}

func TestFinanceMetrics_WriteRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		routeName  string
		httpStatus int
		wantOp     string
		wantStatus string
	}{
		{
			name:       "create_transaction 201 → success",
			method:     http.MethodPost,
			path:       "/finance/transactions",
			routeName:  "create_transaction",
			httpStatus: http.StatusCreated,
			wantOp:     "create_transaction",
			wantStatus: "success",
		},
		{
			name:       "pay_invoice 422 → declined",
			method:     http.MethodPatch,
			path:       "/finance/invoices/1/pay",
			routeName:  "pay_invoice",
			httpStatus: http.StatusUnprocessableEntity,
			wantOp:     "pay_invoice",
			wantStatus: "declined",
		},
		{
			name:       "refund_transaction 500 → technical_error",
			method:     http.MethodPost,
			path:       "/finance/transactions/1/refund",
			routeName:  "refund_transaction",
			httpStatus: http.StatusInternalServerError,
			wantOp:     "refund_transaction",
			wantStatus: "technical_error",
		},
		{
			name:       "update_transaction PUT → counted (R8)",
			method:     http.MethodPut,
			path:       "/finance/transactions/1",
			routeName:  "update_transaction",
			httpStatus: http.StatusOK,
			wantOp:     "update_transaction",
			wantStatus: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, fm := newTestBusinessMetrics(t)
			app := buildApp(bm, tt.method, tt.path, tt.httpStatus, tt.routeName)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			counter := fm.GetCounter("financial_operation_total")
			require.NotNil(t, counter, "financial_operation_total must be registered")

			values := counter.GetValues()
			require.Len(t, values, 1)
			assert.Equal(t, int64(1), values[0].Value)
			assert.True(t, hasStringField(values[0].Fields, "operation", tt.wantOp),
				"expected operation=%s", tt.wantOp)
			assert.True(t, hasStringField(values[0].Fields, "status", tt.wantStatus),
				"expected status=%s", tt.wantStatus)
		})
	}
}

func TestFinanceMetrics_UnnamedRoute_OperationUnknown(t *testing.T) {
	bm, fm := newTestBusinessMetrics(t)
	app := buildApp(bm, http.MethodPost, "/finance/transactions", http.StatusCreated, "")

	req := httptest.NewRequest(http.MethodPost, "/finance/transactions", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	counter := fm.GetCounter("financial_operation_total")
	require.NotNil(t, counter)

	values := counter.GetValues()
	require.Len(t, values, 1, "counter must increment even for unnamed routes")
	assert.True(t, hasStringField(values[0].Fields, "operation", "unknown"),
		"unnamed route must produce operation=unknown")
}

func TestFinanceMetrics_GetRequest_NotCounted(t *testing.T) {
	bm, fm := newTestBusinessMetrics(t)
	app := buildApp(bm, http.MethodGet, "/finance/transactions", http.StatusOK, "list_transactions")

	req := httptest.NewRequest(http.MethodGet, "/finance/transactions", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	counter := fm.GetCounter("financial_operation_total")
	if counter != nil {
		assert.Empty(t, counter.GetValues(), "GET request must not increment the counter")
	}
}

func TestFinanceMetrics_IdempotencyKey_PropagatedToSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	ctx, span := tracer.Start(context.Background(), "test-span")

	bm, _ := newTestBusinessMetrics(t)
	app := buildAppWithContext(bm, http.MethodPost, "/finance/transactions", http.StatusCreated, "create_transaction", ctx)

	req := httptest.NewRequest(http.MethodPost, "/finance/transactions", nil)
	req.Header.Set("Idempotency-Key", "abc-123")

	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected exactly one exported span")

	found := false
	for _, attr := range spans[0].Attributes {
		if string(attr.Key) == "idempotency.key" && attr.Value.AsString() == "abc-123" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected idempotency.key=abc-123 attribute on span")
}

func TestFinanceMetrics_IdempotencyKey_AbsentWhenHeaderMissing(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	ctx, span := tracer.Start(context.Background(), "test-span")

	bm, _ := newTestBusinessMetrics(t)
	app := buildAppWithContext(bm, http.MethodPost, "/finance/transactions", http.StatusCreated, "create_transaction", ctx)

	req := httptest.NewRequest(http.MethodPost, "/finance/transactions", nil)
	// No Idempotency-Key header

	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	for _, attr := range spans[0].Attributes {
		assert.NotEqual(t, "idempotency.key", string(attr.Key),
			"idempotency.key must not be set when header is absent")
	}
}

func TestFinanceMetrics_NilBusinessMetrics_Passthrough(t *testing.T) {
	app := buildApp(nil, http.MethodPost, "/finance/transactions", http.StatusCreated, "create_transaction")

	req := httptest.NewRequest(http.MethodPost, "/finance/transactions", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusCreated, resp.StatusCode,
		"nil BusinessMetrics must pass through without panic")
}

func TestFinanceMetrics_ErrorResponses_AreStillCounted(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus string
	}{
		{
			name:       "fiber error becomes declined",
			err:        fiber.NewError(http.StatusUnprocessableEntity, "validation failed"),
			wantStatus: "declined",
		},
		{
			name:       "generic error becomes technical_error",
			err:        assert.AnError,
			wantStatus: "technical_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, fm := newTestBusinessMetrics(t)
			app := buildAppReturningError(bm, http.MethodPost, "/finance/transactions", tt.err, "create_transaction")

			req := httptest.NewRequest(http.MethodPost, "/finance/transactions", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			counter := fm.GetCounter("financial_operation_total")
			require.NotNil(t, counter)

			values := counter.GetValues()
			require.Len(t, values, 1)
			assert.True(t, hasStringField(values[0].Fields, "status", tt.wantStatus),
				"expected status=%s", tt.wantStatus)
		})
	}
}
