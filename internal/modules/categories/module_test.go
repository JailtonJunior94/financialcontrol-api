package categories_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/routes"
	jwtmocks "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt/mocks"
)

func TestRegisterHTTP_RegistersAuthenticatedRoutes(t *testing.T) {
	parser := jwtmocks.NewParser(t)

	module := &categories.Module{
		CategoryHandler: handlers.NewCategoryHandler(nil, nil, nil, nil, nil),
	}

	app := fiber.New()
	module.RegisterHTTP(app, parser)

	routesList := app.GetRoutes(true)
	got := make(map[string]map[string]bool, len(routesList))
	for _, r := range routesList {
		if got[r.Path] == nil {
			got[r.Path] = map[string]bool{}
		}
		got[r.Path][r.Method] = true
	}

	cases := []struct {
		path    string
		methods []string
	}{
		{"/categories", []string{fiber.MethodGet, fiber.MethodPost}},
		{"/categories/:id", []string{fiber.MethodGet, fiber.MethodPut, fiber.MethodDelete}},
	}
	for _, tc := range cases {
		for _, m := range tc.methods {
			require.Truef(t, got[tc.path][m], "expected %s %s registered", m, tc.path)
		}
	}
}

func TestRegisterCategoryRoutes_RequiresAuthentication(t *testing.T) {
	parser := jwtmocks.NewParser(t)
	parser.EXPECT().Parse(mock.Anything, mock.AnythingOfType("string")).Maybe()

	app := fiber.New()
	handler := handlers.NewCategoryHandler(nil, nil, nil, nil, nil)
	routes.RegisterCategoryRoutes(app, handler, parser)

	cases := []struct {
		method string
		path   string
	}{
		{fiber.MethodGet, "/categories"},
		{fiber.MethodGet, "/categories/some-id"},
		{fiber.MethodPost, "/categories"},
		{fiber.MethodPut, "/categories/some-id"},
		{fiber.MethodDelete, "/categories/some-id"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equalf(t, fiber.StatusUnauthorized, resp.StatusCode, "%s %s should require auth", tc.method, tc.path)
	}
}
