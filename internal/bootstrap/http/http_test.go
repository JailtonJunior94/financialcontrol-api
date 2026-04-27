package http_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"
	cardshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/http"
	cataloghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	invoicinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"

	"github.com/stretchr/testify/require"
)

// stubModule returns an *identity.Module with nil handlers — safe for route-registration
// tests that never invoke any handler.
func stubModule() *identity.Module { return &identity.Module{} }

func TestNewAppRegistersAPIBootstrapRoutes(t *testing.T) {
	app := bootstraphttp.NewApp(&bootstrapcontainer.Container{
		IdentityModule:        stubModule(),
		TransactionController: &transactionshttp.TransactionController{},
		BillController:        &billinghttp.BillController{},
		FlagController:        &cataloghttp.FlagController{},
		CardController:        &cardshttp.CardController{},
		InvoiceController:     &invoicinghttp.InvoiceController{},
		CategoryController:    &cataloghttp.CategoryController{},
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
		IdentityModule:        stubModule(),
		TransactionController: &transactionshttp.TransactionController{},
		BillController:        &billinghttp.BillController{},
		FlagController:        &cataloghttp.FlagController{},
		CardController:        &cardshttp.CardController{},
		InvoiceController:     &invoicinghttp.InvoiceController{},
		CategoryController:    &cataloghttp.CategoryController{},
	})

	require.NotNil(t, app)
	require.Equal(t, "Development", config.Environment)
}
