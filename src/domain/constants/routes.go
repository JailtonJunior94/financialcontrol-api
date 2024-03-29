package constants

const (
	Token = "/token"

	Bills            = "/bills"
	BillDetail       = "/bills/:id"
	BillId           = "/bills/:billid"
	BillsIdAndItemId = "/bills/:billid/items/:id"

	Flags = "/flags"

	Transactions           = "/transactions"
	TransactionDetail      = "/transactions/:id"
	TransactionIdAndItemId = "/transactions/:transactionid/items/:id"
	TransactionId          = "/transactions/:transactionid"
	TransactionClone       = "/transactions/:transactionid/clone"

	Users = "/users"

	Cards  = "/cards"
	CardId = "/cards/:id"

	Invoices           = "/cards/:id/invoices"
	InvoicesImport     = "/cards/:id/invoices-import"
	InvoicesCategories = "/cards/:id/categories"
	InvoicesById       = "/cards/:cardid/invoices/:id"

	Categories = "/categories"
)
