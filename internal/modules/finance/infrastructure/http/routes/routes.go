package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func RegisterFinanceRoutes(
	router fiber.Router,
	transaction *handlers.TransactionHandler,
	invoice *handlers.InvoiceHandler,
	installment *handlers.InstallmentHandler,
	summary *handlers.SummaryHandler,
	parser pkgjwt.Parser,
) {
	protected := authmiddleware.Protected(parser)

	router.Post(pkgroutes.FinanceTransactions, protected, transaction.Create)
	router.Get(pkgroutes.FinanceTransactions, protected, transaction.List)
	router.Get(pkgroutes.FinanceTransactionID, protected, transaction.Get)
	router.Put(pkgroutes.FinanceTransactionID, protected, transaction.Update)
	router.Delete(pkgroutes.FinanceTransactionID, protected, transaction.Delete)
	router.Post(pkgroutes.FinanceTransactionRefund, protected, transaction.Refund)

	router.Get(pkgroutes.FinanceInvoices, protected, invoice.List)
	router.Get(pkgroutes.FinanceInvoiceID, protected, invoice.Get)
	router.Patch(pkgroutes.FinanceInvoicePay, protected, invoice.Pay)

	router.Post(pkgroutes.FinanceInstallmentAnticipate, protected, installment.Anticipate)

	router.Get(pkgroutes.FinanceSummary, protected, summary.Get)
}
