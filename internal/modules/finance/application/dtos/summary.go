package dtos

// MonthlySummaryResponse carries the 7 aggregate buckets plus the computed balance (RF-20/RF-51).
// All monetary values are serialised as decimal strings (decisão H3.a).
type MonthlySummaryResponse struct {
	TotalIncome               string `json:"total_income"`
	TotalExpense              string `json:"total_expense"`
	TotalRefundsIn            string `json:"total_refunds_in"`
	TotalRefundsOut           string `json:"total_refunds_out"`
	TotalCreditPurchasesMonth string `json:"total_credit_purchases_month"`
	TotalInvoicesOpen         string `json:"total_invoices_open"`
	TotalInvoicesPaid         string `json:"total_invoices_paid"`
	Balance                   string `json:"balance"`
}
