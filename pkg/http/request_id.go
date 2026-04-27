package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID returns the X-Request-ID header value or generates a new UUID v4 when absent.
func RequestID(c *fiber.Ctx) string {
	if id := c.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}
