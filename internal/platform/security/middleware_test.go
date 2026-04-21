package security

import (
	"net/http/httptest"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestProtectedReturnsUnauthorizedForMissingToken(t *testing.T) {
	config.JwtSecret = "test-secret"

	app := fiber.New()
	app.Get("/protected", Protected(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
