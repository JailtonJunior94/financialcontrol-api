package modules_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	cardshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categorieshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financehandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
)

func stubCardsModule() *cards.Module {
	return &cards.Module{
		CardHandler: cardshandlers.NewCardHandler(nil, nil, nil, nil, nil),
		FlagHandler: cardshandlers.NewFlagHandler(nil),
	}
}

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

func TestRegistrationsExposeExpectedFoundationModules(t *testing.T) {
	registrations := modules.Registrations()
	require.Empty(t, registrations, "all legacy module registrations removed in task 9.0")
}

// TestRegisterHTTP_WiresAllModuleGroups verifies that RegisterHTTP delegates to all
// four module groups. The apiV1Router in internal/bootstrap/http/router.go calls this
// function — changes here break the registration contract for the entire API surface.
func TestRegisterHTTP_WiresAllModuleGroups(t *testing.T) {
	app := fiber.New()
	group := app.Group("/api/v1")

	container := &bootstrapcontainer.Container{
		IdentityModule:   &identity.Module{},
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
	}

	modules.RegisterHTTP(group, container)

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, r := range routes {
		if routeIndex[r.Path] == nil {
			routeIndex[r.Path] = map[string]bool{}
		}
		routeIndex[r.Path][r.Method] = true
	}

	// Spot-check one route from each module to confirm all groups are wired.
	require.True(t, routeIndex["/api/v1/token"]["POST"], "identity module: POST /api/v1/token")
	require.True(t, routeIndex["/api/v1/cards"]["POST"], "cards module: POST /api/v1/cards")
	require.True(t, routeIndex["/api/v1/categories"]["GET"], "categories module: GET /api/v1/categories")
	require.True(t, routeIndex["/api/v1/finance/transactions"]["POST"], "finance module: POST /api/v1/finance/transactions")
}
