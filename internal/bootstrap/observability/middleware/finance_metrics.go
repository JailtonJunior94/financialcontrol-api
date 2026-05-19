package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/metrics"
)

// FinanceMetrics returns a Fiber middleware that records financial operation metrics
// for write methods (POST/PUT/PATCH/DELETE) on the /api/v1/finance group.
// Nil BusinessMetrics is safe: the middleware degrades to a pass-through.
func FinanceMetrics(b *metrics.BusinessMetrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if b == nil || !isWriteMethod(c.Method()) {
			return c.Next()
		}

		err := c.Next()
		if err != nil {
			applyErrorStatus(c, err)
		}

		op := resolveOperation(c)
		status := resolveStatus(c)
		b.RecordOperation(c.UserContext(), op, status)
		propagateIdempotency(c)

		return err
	}
}

func applyErrorStatus(c *fiber.Ctx, err error) {
	if ferr, ok := err.(*fiber.Error); ok {
		c.Status(ferr.Code)
		return
	}
	c.Status(fiber.StatusInternalServerError)
}

func isWriteMethod(method string) bool {
	switch method {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		return true
	}
	return false
}

func resolveOperation(c *fiber.Ctx) metrics.OperationLabel {
	if r := c.Route(); r != nil && r.Name != "" {
		return metrics.ParseOperationLabel(r.Name)
	}
	return metrics.OperationUnknown
}

func resolveStatus(c *fiber.Ctx) metrics.StatusBucket {
	return metrics.StatusFromHTTP(c.Response().StatusCode())
}

func propagateIdempotency(c *fiber.Ctx) {
	key := c.Get("Idempotency-Key")
	if key == "" {
		return
	}

	span := trace.SpanFromContext(c.UserContext())
	span.SetAttributes(attribute.String("idempotency.key", key))
	slog.Default().InfoContext(c.UserContext(), "idempotency.key propagated",
		slog.String("idempotency.key", key),
	)
}
