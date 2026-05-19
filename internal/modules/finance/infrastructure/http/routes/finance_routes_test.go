package routes_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/routes"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func TestRegisterFinanceRoutes_RegistersAllPaths(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	noop := func(c *fiber.Ctx) error { return c.Next() }

	routes.RegisterFinanceRoutes(
		app,
		&handlers.TransactionHandler{},
		&handlers.InvoiceHandler{},
		&handlers.InstallmentHandler{},
		&handlers.SummaryHandler{},
		noop,
	)

	paths := make(map[string]bool)
	for _, r := range app.GetRoutes() {
		paths[r.Path] = true
	}

	for _, want := range []string{
		pkgroutes.FinanceTransactions,
		pkgroutes.FinanceTransactionID,
		pkgroutes.FinanceTransactionRefund,
		pkgroutes.FinanceInvoices,
		pkgroutes.FinanceInvoiceID,
		pkgroutes.FinanceInvoicePay,
		pkgroutes.FinanceInstallmentAnticipate,
		pkgroutes.FinanceSummary,
	} {
		assert.Truef(t, paths[want], "route %s must be registered", want)
	}
}
