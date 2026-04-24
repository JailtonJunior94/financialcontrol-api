package catalog_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog"
	cataloghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegistrationExposesCatalogHTTPRoutes(t *testing.T) {
	registration := catalog.Registration()

	require.Equal(t, "catalog", registration.Name)
	require.NotNil(t, registration.RegisterHTTP)
	require.Nil(t, registration.RegisterCLI)

	app := fiber.New()
	registration.RegisterHTTP(app, &container.Container{
		FlagController:     &cataloghttp.FlagController{},
		CategoryController: &cataloghttp.CategoryController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}

		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/flags"][fiber.MethodGet])
	require.True(t, routeIndex["/categories"][fiber.MethodGet])
}
