package invoicing

import "context"

// InvoiceChangedPublisher is a port that the invoicing module uses to notify
// other modules that an invoice has been updated. Implementations are
// registered in the bootstrap and injected into the invoicing application
// service; no concrete repository of another module is accessed directly.
type InvoiceChangedPublisher interface {
	PublishInvoiceChanged(ctx context.Context, invoiceID string) error
}
