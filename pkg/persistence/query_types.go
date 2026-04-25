package persistence

import "time"

// BillQuery is the read model returned by billing repository queries used for
// planning and reporting. It is defined here so both the billing module and the
// planning module can reference it without cross-module imports.
type BillQuery struct {
	ID    string
	Date  time.Time
	Items []BillItemQuery
}

type BillItemQuery struct {
	ID          string
	Description string
	Total       float64
}

// InvoiceQuery is the minimal projection used to aggregate invoice data
// across cards for planning purposes.
type InvoiceQuery struct {
	Date        time.Time `db:"Date"`
	Description string    `db:"Description"`
	Total       float64   `db:"Total"`
}

// InvoiceRead is the full read model for a single invoice with its line items.
type InvoiceRead struct {
	ID    string
	Date  time.Time
	Total float64
	Items []InvoiceItemRead
}

type InvoiceItemRead struct {
	ID               string
	PurchaseDate     time.Time
	Description      string
	TotalAmount      float64
	Installment      int
	InstallmentValue float64
	Tags             string
	Category         CategoryRead
}

type CategoryRead struct {
	ID   string
	Name string
}

// TransactionQuery is the minimal projection used to locate a transaction by
// date and card for synchronisation with invoices.
type TransactionQuery struct {
	ID            string `db:"Id"`
	TransactionID string `db:"TransactionId"`
	UserID        string `db:"UserId"`
}
