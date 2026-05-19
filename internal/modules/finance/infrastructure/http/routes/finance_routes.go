package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func RegisterFinanceRoutes(
	router fiber.Router,
	transaction *handlers.TransactionHandler,
	invoice *handlers.InvoiceHandler,
	installment *handlers.InstallmentHandler,
	summary *handlers.SummaryHandler,
	protected fiber.Handler,
) {
	router.Post(pkgroutes.FinanceTransactions, protected, transaction.Create).Name("create_transaction")
	router.Get(pkgroutes.FinanceTransactions, protected, transaction.List)
	router.Get(pkgroutes.FinanceTransactionID, protected, transaction.Get)
	router.Put(pkgroutes.FinanceTransactionID, protected, transaction.Update).Name("update_transaction")
	router.Delete(pkgroutes.FinanceTransactionID, protected, transaction.Delete).Name("delete_transaction")
	router.Post(pkgroutes.FinanceTransactionRefund, protected, transaction.Refund).Name("refund_transaction")

	router.Get(pkgroutes.FinanceInvoices, protected, invoice.List)
	router.Get(pkgroutes.FinanceInvoiceID, protected, invoice.Get)
	router.Patch(pkgroutes.FinanceInvoicePay, protected, invoice.Pay).Name("pay_invoice")

	router.Post(pkgroutes.FinanceInstallmentAnticipate, protected, installment.Anticipate).Name("anticipate_installment")

	router.Get(pkgroutes.FinanceSummary, protected, summary.Get)
}
