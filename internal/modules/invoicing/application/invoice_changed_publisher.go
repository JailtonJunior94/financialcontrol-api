package invoicingapp

import (
	"context"
	"fmt"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/pkg/events"
)

// InvoiceChangedEventPublisher adapts the platform Dispatcher to the
// InvoiceChangedPublisher port. It is constructed in bootstrap and injected
// wherever the invoicing service needs to publish the event.
type InvoiceChangedEventPublisher struct {
	dispatcher platformevents.EventDispatcher
}

// NewInvoiceChangedEventPublisher returns a publisher backed by the given dispatcher.
func NewInvoiceChangedEventPublisher(dispatcher platformevents.EventDispatcher) *InvoiceChangedEventPublisher {
	return &InvoiceChangedEventPublisher{dispatcher: dispatcher}
}

// PublishInvoiceChanged fires the invoice_changed event with the given payload.
func (p *InvoiceChangedEventPublisher) PublishInvoiceChanged(_ context.Context, payload invoicingdomain.InvoiceChangedPayload) error {
	if payload.InvoiceID == "" {
		return fmt.Errorf("invoice_changed publisher: InvoiceID must not be empty")
	}
	p.dispatcher.Dispatch(invoicingdomain.NewInvoiceChangedEvent(payload))
	return nil
}
