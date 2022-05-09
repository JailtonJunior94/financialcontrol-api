package constants

const (
	Token                  = "/token"
	Bills                  = "/bills"
	BillDetail             = "/bills/:id"
	BillId                 = "/bills/:billid"
	BillsIdAndItemId       = "/bills/:billid/items/:id"
	Flags                  = "/flags"
	Categories             = "/categories"
	Transactions           = "/transactions"
	TransactionDetail      = "/transactions/:id"
	TransactionIdAndItemId = "/transactions/:transactionid/items/:id"
	TransactionId          = "/transactions/:transactionid"
	TransactionClone       = "/transactions/:transactionid/clone"
	Users                  = "/users"
	Cards                  = "/cards"
	CardId                 = "/cards/:id"
	Invoices               = "/invoices"
	InvoicesImport         = "/invoices-import"
	InvoicesCategories     = "/invoices/:id/categories"
	InvoicesById           = "/invoices/:id"
	InvoicesItems          = "/invoices/:id/items"
)
