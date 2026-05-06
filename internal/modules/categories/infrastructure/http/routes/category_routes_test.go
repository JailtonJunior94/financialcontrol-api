package routes_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

// TestCategoryIdRouteSegment verifies the dynamic /categories/:id segment
// resolves correctly for all CRUD verbs.
func TestCategoryIdRouteSegment(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	hits := map[string]bool{}
	app.Get(pkgroutes.Categories, func(c *fiber.Ctx) error { hits["list"] = true; return c.SendStatus(fiber.StatusOK) })
	app.Get(pkgroutes.CategoryId, func(c *fiber.Ctx) error { hits["get"] = true; return c.SendStatus(fiber.StatusOK) })
	app.Post(pkgroutes.Categories, func(c *fiber.Ctx) error { hits["create"] = true; return c.SendStatus(fiber.StatusCreated) })
	app.Put(pkgroutes.CategoryId, func(c *fiber.Ctx) error { hits["update"] = true; return c.SendStatus(fiber.StatusOK) })
	app.Delete(pkgroutes.CategoryId, func(c *fiber.Ctx) error { hits["delete"] = true; return c.SendStatus(fiber.StatusNoContent) })

	cases := []struct {
		method, path, key string
		status            int
	}{
		{"GET", "/categories", "list", fiber.StatusOK},
		{"GET", "/categories/abc", "get", fiber.StatusOK},
		{"POST", "/categories", "create", fiber.StatusCreated},
		{"PUT", "/categories/abc", "update", fiber.StatusOK},
		{"DELETE", "/categories/abc", "delete", fiber.StatusNoContent},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		res, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, c.status, res.StatusCode, c.key)
		assert.True(t, hits[c.key], "%s handler must be called", c.key)
	}
}
