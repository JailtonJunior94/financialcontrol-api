package invoicing_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing"
	invoicinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegistrationExposesInvoicingHTTPRoutes(t *testing.T) {
	registration := invoicing.Registration()

	require.Equal(t, "invoicing", registration.Name)
	require.NotNil(t, registration.RegisterHTTP)
	require.Nil(t, registration.RegisterCLI)

	app := fiber.New()
	registration.RegisterHTTP(app, &container.Container{
		InvoiceController: &invoicinghttp.InvoiceController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}

		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/invoices"][fiber.MethodGet])
	require.True(t, routeIndex["/invoices"][fiber.MethodPost])
	require.True(t, routeIndex["/invoices/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/invoices/:id"][fiber.MethodPatch])
	require.True(t, routeIndex["/invoices/:id/items"][fiber.MethodPut])
	require.True(t, routeIndex["/invoices/:id/items"][fiber.MethodDelete])
	require.True(t, routeIndex["/invoices-import"][fiber.MethodPost])
	require.True(t, routeIndex["/invoices/:id/categories"][fiber.MethodGet])
}
