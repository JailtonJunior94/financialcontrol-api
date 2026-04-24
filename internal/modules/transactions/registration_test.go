package transactions_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegistrationExposesTransactionHTTPRoutes(t *testing.T) {
	registration := transactions.Registration()

	require.Equal(t, "transactions", registration.Name)
	require.NotNil(t, registration.RegisterHTTP)
	require.Nil(t, registration.RegisterCLI)

	app := fiber.New()
	registration.RegisterHTTP(app, &container.Container{
		TransactionController: &transactionshttp.TransactionController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}

		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/transactions"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions/:transactionid"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:transactionid/clone"][fiber.MethodPost])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodGet])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodPut])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodPatch])
	require.True(t, routeIndex["/transactions/:transactionid/items/:id"][fiber.MethodDelete])
}
