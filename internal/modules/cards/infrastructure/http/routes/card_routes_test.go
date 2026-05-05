package routes_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

// TestCardFlagsResolvedBeforeCardID verifies that GET /cards/flags matches its
// own route and never falls through to the dynamic /cards/:id segment.
func TestCardFlagsResolvedBeforeCardID(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	var flagsHit, idHit bool

	app.Get(pkgroutes.CardFlags, func(c *fiber.Ctx) error {
		flagsHit = true
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get(pkgroutes.CardId, func(c *fiber.Ctx) error {
		idHit = true
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/cards/flags", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, flagsHit, "/cards/flags handler must be called")
	assert.False(t, idHit, "/cards/:id must NOT be called for /cards/flags")
}
