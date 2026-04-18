package routes_test

import (
	"testing"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/controllers"

	"github.com/stretchr/testify/require"
)

func TestRegisterPreservesHTTPContract(t *testing.T) {
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
		"/api/v1/flags":                                 {"GET"},
		"/api/v1/cards":                                 {"GET", "POST"},
		"/api/v1/cards/:id":                             {"GET", "PUT", "DELETE"},
		"/api/v1/invoices":                              {"GET", "POST"},
		"/api/v1/invoices/:id":                          {"GET", "PATCH"},
		"/api/v1/invoices/:id/items":                    {"PUT", "DELETE"},
		"/api/v1/invoices-import":                       {"POST"},
		"/api/v1/invoices/:id/categories":               {"GET"},
		"/api/v1/categories":                            {"GET"},
	}

	for path, methods := range expected {
		for _, method := range methods {
			require.Truef(t, routeIndex[path][method], "expected route %s %s", method, path)
		}
	}
}
