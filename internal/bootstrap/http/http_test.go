package http_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/controllers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"

	"github.com/stretchr/testify/require"
)

func TestNewAppRegistersAPIBootstrapRoutes(t *testing.T) {
	app := bootstraphttp.NewApp(&bootstrapcontainer.Container{
		UserController:        &controllers.UserController{},
		AuthController:        &controllers.AuthController{},
		TransactionController: &controllers.TransactionController{},
		BillController:        &controllers.BillController{},
		FlagController:        &controllers.FlagController{},
		CardController:        &controllers.CardController{},
		InvoiceController:     &controllers.InvoiceController{},
		CategoryController:    &controllers.CategoryController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	require.True(t, routeIndex["/api/v1/token"]["POST"])
	require.True(t, routeIndex["/api/v1/users"]["POST"])
	require.True(t, routeIndex["/api/v1/me"]["GET"])
	require.True(t, routeIndex["/api/v1/transactions"]["GET"])
	require.True(t, routeIndex["/api/v1/transactions"]["POST"])
	require.True(t, routeIndex["/api/v1/bills"]["GET"])
	require.True(t, routeIndex["/api/v1/cards"]["POST"])
	require.True(t, routeIndex["/api/v1/invoices-import"]["POST"])
	require.True(t, routeIndex["/api/v1/categories"]["GET"])
}

func TestNewAppBootstrapsWithRuntimeConfigLoaded(t *testing.T) {
	t.Setenv("ENVIRONMENT", "Development")

	err := config.Load()
	require.NoError(t, err)

	app := bootstraphttp.NewApp(&bootstrapcontainer.Container{
		UserController:        &controllers.UserController{},
		AuthController:        &controllers.AuthController{},
		TransactionController: &controllers.TransactionController{},
		BillController:        &controllers.BillController{},
		FlagController:        &controllers.FlagController{},
		CardController:        &controllers.CardController{},
		InvoiceController:     &controllers.InvoiceController{},
		CategoryController:    &controllers.CategoryController{},
	})

	require.NotNil(t, app)
	require.Equal(t, "Development", config.Environment)
}
