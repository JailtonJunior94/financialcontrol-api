package planning

import (
	"context"
	"time"
)

// BillingReadPort is a read port consumed by the planning module to obtain
// monthly billing data without depending on concrete billing repositories.
type BillingReadPort interface {
	GetMonthlyBills(ctx context.Context, referenceDate time.Time) (*MonthlyBillsReadModel, error)
}

// InvoicingReadPort is a read port consumed by the planning module to obtain
// monthly invoicing data without depending on concrete invoicing repositories.
type InvoicingReadPort interface {
	GetMonthlyInvoices(ctx context.Context, referenceDate time.Time) (*MonthlyInvoicesReadModel, error)
}

// SyncPort is a port consumed by the planning module to trigger operational
// synchronisation of transactions and invoices without coupling to concrete use cases.
type SyncPort interface {
	Sync() error
}

// MonthlyBillsReadModel carries aggregated billing data for a reference month.
type MonthlyBillsReadModel struct {
	ReferenceDate time.Time
	Items         []BillReadItem
}

// BillReadItem represents a single bill line used for planning reports.
type BillReadItem struct {
	ID          string
	Description string
	Total       float64
}

// MonthlyInvoicesReadModel carries aggregated invoicing data for a reference month.
type MonthlyInvoicesReadModel struct {
	ReferenceDate time.Time
	Items         []InvoiceReadItem
}

// InvoiceReadItem represents a single invoice line used for planning reports.
// Total maps to the installment value for the reference month.
// Tags is the budget category tag used to group spending.
// Category is the sub-category name used for drill-down reports.
type InvoiceReadItem struct {
	ID          string
	Description string
	Total       float64
	Date        time.Time
	Tags        string
	Category    string
}
