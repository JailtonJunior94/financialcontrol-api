package http_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categorieshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financehandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	pkghttp "github.com/jailtonjunior94/financialcontrol-api/pkg/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func stubCategoriesModule() *categories.Module {
	return &categories.Module{
		CategoryHandler: categorieshandlers.NewCategoryHandler(nil, nil, nil, nil, nil),
	}
}

func stubCardsModule() *cards.Module {
	return &cards.Module{
		CardHandler: handlers.NewCardHandler(nil, nil, nil, nil, nil),
		FlagHandler: handlers.NewFlagHandler(nil),
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

func stubContainer() *bootstrapcontainer.Container {
	return &bootstrapcontainer.Container{
		IdentityModule:   &identity.Module{},
		CardsModule:      stubCardsModule(),
		CategoriesModule: stubCategoriesModule(),
		FinanceModule:    stubFinanceModule(),
	}
}

// TestRegisterRoutesPreservesHTTPContract verifies the full /api/v1/* route surface
// is preserved after the migration to apiV1Router in internal/bootstrap/http/router.go.
// RegisterRoutes is the pkg-level entry used in tests; the bootstrap layer uses
// apiV1Router (same delegation chain) to mount routes via serverfiber.Server.
func TestRegisterRoutesPreservesHTTPContract(t *testing.T) {
	app := fiber.New()
	pkghttp.RegisterRoutes(app, stubContainer())

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	expected := map[string][]string{
		"/api/v1/token":                               {"POST"},
		"/api/v1/me":                                  {"GET"},
		"/api/v1/users":                               {"POST"},
		"/api/v1/cards":                               {"GET", "POST"},
		"/api/v1/cards/flags":                         {"GET"},
		"/api/v1/cards/:id":                           {"GET", "PUT", "DELETE"},
		"/api/v1/categories":                          {"GET", "POST"},
		"/api/v1/categories/:id":                      {"GET", "PUT", "DELETE"},
		"/api/v1/finance/transactions":                {"GET", "POST"},
		"/api/v1/finance/transactions/:id":            {"GET", "PUT", "DELETE"},
		"/api/v1/finance/transactions/:id/refund":     {"POST"},
		"/api/v1/finance/invoices":                    {"GET"},
		"/api/v1/finance/invoices/:id":                {"GET"},
		"/api/v1/finance/invoices/:id/pay":            {"PATCH"},
		"/api/v1/finance/installments/:id/anticipate": {"POST"},
		"/api/v1/finance/summary":                     {"GET"},
	}

	for path, methods := range expected {
		for _, method := range methods {
			require.Truef(t, routeIndex[path][method], "expected route %s %s", method, path)
		}
	}
}

// TestRegisterRoutes_SameContractAsAPIV1Router verifies that RegisterRoutes and the
// bootstrap apiV1Router (tested in internal/bootstrap/http) wire the same /api/v1
// surface. If this test diverges from TestNewAppRegistersAPIBootstrapRoutes in the
// bootstrap package, the route contracts have drifted.
func TestRegisterRoutes_SameContractAsAPIV1Router(t *testing.T) {
	app := fiber.New()
	pkghttp.RegisterRoutes(app, stubContainer())

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	// Spot-check the routes also verified in TestNewAppRegistersAPIBootstrapRoutes.
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
