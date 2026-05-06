package http_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categorieshandlers "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	invoicinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
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

func TestRegisterRoutesPreservesHTTPContract(t *testing.T) {
	app := fiber.New()
	pkghttp.RegisterRoutes(app, &bootstrapcontainer.Container{
		IdentityModule:        &identity.Module{},
		TransactionController: &transactionshttp.TransactionController{},
		BillController:        &billinghttp.BillController{},
		CardsModule:           stubCardsModule(),
		CategoriesModule:      stubCategoriesModule(),
		InvoiceController:     &invoicinghttp.InvoiceController{},
	})

	routes := app.GetRoutes(true)
	routeIndex := make(map[string]map[string]bool, len(routes))
	for _, route := range routes {
		if routeIndex[route.Path] == nil {
			routeIndex[route.Path] = map[string]bool{}
		}
		routeIndex[route.Path][route.Method] = true
	}

	expected := map[string][]string{
		"/api/v1/token":                                 {"POST"},
		"/api/v1/me":                                    {"GET"},
		"/api/v1/users":                                 {"POST"},
		"/api/v1/transactions":                          {"GET", "POST"},
		"/api/v1/transactions/:id":                      {"GET"},
		"/api/v1/transactions/:transactionid":           {"POST"},
		"/api/v1/transactions/:transactionid/clone":     {"POST"},
		"/api/v1/transactions/:transactionid/items/:id": {"GET", "PUT", "PATCH", "DELETE"},
		"/api/v1/bills":                                 {"GET", "POST"},
		"/api/v1/bills/:id":                             {"GET"},
		"/api/v1/bills/:billid":                         {"POST"},
		"/api/v1/bills/:billid/items/:id":               {"GET", "PUT", "DELETE"},
		"/api/v1/cards":                                 {"GET", "POST"},
		"/api/v1/cards/flags":                           {"GET"},
		"/api/v1/cards/:id":                             {"GET", "PUT", "DELETE"},
		"/api/v1/invoices":                              {"GET", "POST"},
		"/api/v1/invoices/:id":                          {"GET", "PATCH"},
		"/api/v1/invoices/:id/items":                    {"PUT", "DELETE"},
		"/api/v1/invoices-import":                       {"POST"},
		"/api/v1/invoices/:id/categories":               {"GET"},
		"/api/v1/categories":                            {"GET", "POST"},
		"/api/v1/categories/:id":                        {"GET", "PUT", "DELETE"},
	}

	for path, methods := range expected {
		for _, method := range methods {
			require.Truef(t, routeIndex[path][method], "expected route %s %s", method, path)
		}
	}
}
