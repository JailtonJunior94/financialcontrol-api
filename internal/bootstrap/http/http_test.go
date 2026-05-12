package http_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categorieshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financehandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"

	"github.com/stretchr/testify/require"
)

func stubCardsModule() *cards.Module {
	return &cards.Module{
		CardHandler: handlers.NewCardHandler(nil, nil, nil, nil, nil),
		FlagHandler: handlers.NewFlagHandler(nil),
	}
}

// stubModule returns an *identity.Module with nil handlers — safe for route-registration
// tests that never invoke any handler.
func stubModule() *identity.Module { return &identity.Module{} }

func stubCategoriesModule() *categories.Module {
	return &categories.Module{
		CategoryHandler: categorieshandlers.NewCategoryHandler(nil, nil, nil, nil, nil),
	}
}

func stubFinanceModule() *finance.Module {
	return &finance.Module{
		TransactionHandler: financehandlers.NewTransactionHandler(nil, nil, nil, nil, nil, nil),
		InvoiceHandler:     financehandlers.NewInvoiceHandler(nil, nil, nil),
		InstallmentHandler: financehandlers.NewInstallmentHandler(nil),
		SummaryHandler:     financehandlers.NewSummaryHandler(nil),
	}
}

func TestNewAppRegistersAPIBootstrapRoutes(t *testing.T) {
	app := bootstraphttp.NewApp(&bootstrapcontainer.Container{
		IdentityModule:   stubModule(),
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
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
	require.True(t, routeIndex["/api/v1/cards"]["POST"])
	require.True(t, routeIndex["/api/v1/categories"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/transactions"]["POST"])
	require.True(t, routeIndex["/api/v1/finance/transactions"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/invoices"]["GET"])
	require.True(t, routeIndex["/api/v1/finance/summary"]["GET"])
}

func TestNewAppBootstrapsWithRuntimeConfigLoaded(t *testing.T) {
	t.Setenv("ENVIRONMENT", "Development")

	err := config.Load()
	require.NoError(t, err)

	app := bootstraphttp.NewApp(&bootstrapcontainer.Container{
		IdentityModule:   stubModule(),
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
	})

	require.NotNil(t, app)
	require.Equal(t, "Development", config.Environment)
}
