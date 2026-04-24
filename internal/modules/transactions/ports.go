package transactions

import (
	"context"
	"time"
)

// InvoiceSyncPort is consumed by the transactions module to synchronise
// transaction values when an invoice is recalculated. It is defined here
// (consumer side) so transactions never imports invoicing repositories.
type InvoiceSyncPort interface {
	SyncTransactionWithInvoice(ctx context.Context, invoiceID string, cardDescription string, userID string, referenceDate time.Time, total float64) error
}
