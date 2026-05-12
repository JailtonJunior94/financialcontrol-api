package routes

const (
	FinanceTransactions          = "/finance/transactions"
	FinanceTransactionID         = "/finance/transactions/:id"
	FinanceTransactionRefund     = "/finance/transactions/:id/refund"
	FinanceInvoices              = "/finance/invoices"
	FinanceInvoiceID             = "/finance/invoices/:id"
	FinanceInvoicePay            = "/finance/invoices/:id/pay"
	FinanceInstallmentAnticipate = "/finance/installments/:id/anticipate"
	FinanceSummary               = "/finance/summary"
)
