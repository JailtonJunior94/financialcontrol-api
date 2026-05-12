package services

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
)

// InvoiceCloser applies lazy on-read closing to open invoices (RF-17).
// The caller is responsible for persisting the invoices that were changed.
type InvoiceCloser struct{}

func NewInvoiceCloser() *InvoiceCloser { return &InvoiceCloser{} }

// CloseIfDue iterates the given invoices, calls Invoice.CloseIfDue(now) on each,
// and returns the slice of invoices that were actually transitioned to closed.
// Idempotent: re-calling with already-closed invoices returns an empty slice.
func (c *InvoiceCloser) CloseIfDue(invoices []*entities.Invoice, now time.Time) []*entities.Invoice {
	var changed []*entities.Invoice
	for _, inv := range invoices {
		if inv.CloseIfDue(now) {
			changed = append(changed, inv)
		}
	}
	return changed
}
