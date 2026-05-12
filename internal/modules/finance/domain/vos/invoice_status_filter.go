package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// InvoiceStatusFilter restricts invoice listing to a specific state.
type InvoiceStatusFilter string

const (
	InvoiceStatusFilterNone   InvoiceStatusFilter = "none"
	InvoiceStatusFilterOpen   InvoiceStatusFilter = "open"
	InvoiceStatusFilterClosed InvoiceStatusFilter = "closed"
	InvoiceStatusFilterPaid   InvoiceStatusFilter = "paid"
)

var validInvoiceStatusFilters = map[InvoiceStatusFilter]struct{}{
	InvoiceStatusFilterNone:   {},
	InvoiceStatusFilterOpen:   {},
	InvoiceStatusFilterClosed: {},
	InvoiceStatusFilterPaid:   {},
}

// ParseInvoiceStatusFilter parses a raw string into an InvoiceStatusFilter.
// An empty string is treated as "none" (no filter applied).
func ParseInvoiceStatusFilter(s string) (InvoiceStatusFilter, error) {
	if s == "" {
		return InvoiceStatusFilterNone, nil
	}
	f := InvoiceStatusFilter(s)
	if _, ok := validInvoiceStatusFilters[f]; !ok {
		return "", domain.ErrInvalidTransactionType
	}
	return f, nil
}

func (f InvoiceStatusFilter) String() string { return string(f) }

// IsActive reports whether the filter actually restricts the result set.
func (f InvoiceStatusFilter) IsActive() bool { return f != InvoiceStatusFilterNone && f != "" }
