package routes_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/routes"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func TestRegisterAuthRoutes_RegistersAllPaths(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	noop := func(c *fiber.Ctx) error { return c.Next() }

	routes.RegisterAuthRoutes(app, &handlers.AuthHandler{}, noop)

	paths := make(map[string]bool)
	for _, r := range app.GetRoutes() {
		paths[r.Path] = true
	}

	assert.True(t, paths[pkgroutes.Token], "POST /token must be registered")
	assert.True(t, paths[pkgroutes.Me], "GET /me must be registered")
}
