package billing_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing"
	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegistrationExposesBillingHTTPRoutes(t *testing.T) {
	registration := billing.Registration()

	require.Equal(t, "billing", registration.Name)
	require.NotNil(t, registration.RegisterHTTP)
	require.Nil(t, registration.RegisterCLI)

	app := fiber.New()
	registration.RegisterHTTP(app, &container.Container{
		BillController: &billinghttp.BillController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}

		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/bills"][fiber.MethodGet])
	require.True(t, routeIndex["/bills"][fiber.MethodPost])
	require.True(t, routeIndex["/bills/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/bills/:billid"][fiber.MethodPost])
	require.True(t, routeIndex["/bills/:billid/items/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/bills/:billid/items/:id"][fiber.MethodPut])
	require.True(t, routeIndex["/bills/:billid/items/:id"][fiber.MethodDelete])
}
