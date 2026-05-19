package finance_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func TestRegisterHTTP_RegistersAuthenticatedRoutes(t *testing.T) {
	module := &finance.Module{
		TransactionHandler: handlers.NewTransactionHandler(nil, nil, nil, nil, nil, nil),
		InvoiceHandler:     handlers.NewInvoiceHandler(nil, nil, nil),
		InstallmentHandler: handlers.NewInstallmentHandler(nil),
		SummaryHandler:     handlers.NewSummaryHandler(nil),
	}

	app := fiber.New()
	module.RegisterHTTP(app, allowAllFinance())

	routes := app.GetRoutes(true)
	got := make(map[string]map[string]bool, len(routes))
	for _, r := range routes {
		if got[r.Path] == nil {
			got[r.Path] = map[string]bool{}
		}
		got[r.Path][r.Method] = true
	}

	cases := []struct {
		path    string
		methods []string
	}{
		{pkgroutes.FinanceTransactions, []string{fiber.MethodGet, fiber.MethodPost}},
		{pkgroutes.FinanceTransactionID, []string{fiber.MethodGet, fiber.MethodPut, fiber.MethodDelete}},
		{pkgroutes.FinanceTransactionRefund, []string{fiber.MethodPost}},
		{pkgroutes.FinanceInvoices, []string{fiber.MethodGet}},
		{pkgroutes.FinanceInvoiceID, []string{fiber.MethodGet}},
		{pkgroutes.FinanceInvoicePay, []string{fiber.MethodPatch}},
		{pkgroutes.FinanceInstallmentAnticipate, []string{fiber.MethodPost}},
		{pkgroutes.FinanceSummary, []string{fiber.MethodGet}},
	}
	for _, tc := range cases {
		for _, method := range tc.methods {
			require.Truef(t, got[tc.path][method], "expected %s %s registered", method, tc.path)
		}
	}
}

func TestRegisterFinanceRoutes_RequiresAuthentication(t *testing.T) {
	app := fiber.New()
	routes := &finance.Module{
		TransactionHandler: handlers.NewTransactionHandler(nil, nil, nil, nil, nil, nil),
		InvoiceHandler:     handlers.NewInvoiceHandler(nil, nil, nil),
		InstallmentHandler: handlers.NewInstallmentHandler(nil),
		SummaryHandler:     handlers.NewSummaryHandler(nil),
	}
	routes.RegisterHTTP(app, rejectAllFinance())

	cases := []struct {
		method string
		path   string
	}{
		{fiber.MethodGet, pkgroutes.FinanceTransactions},
		{fiber.MethodPost, pkgroutes.FinanceTransactions},
		{fiber.MethodGet, "/finance/transactions/some-id"},
		{fiber.MethodPatch, "/finance/invoices/some-id/pay"},
		{fiber.MethodPost, "/finance/installments/some-id/anticipate"},
		{fiber.MethodGet, pkgroutes.FinanceSummary},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equalf(t, fiber.StatusUnauthorized, resp.StatusCode, "%s %s should require auth", tc.method, tc.path)
	}
}

func allowAllFinance() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.Next() }
}

func rejectAllFinance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
}
