package invoicingapp

import (
	"context"
	"fmt"

	platformevents "github.com/jailtonjunior94/financialcontrol-api/internal/platform/events"
)

// EventDispatcherPort is the minimal interface this publisher requires from
// the event dispatcher so that the invoicing application layer does not
// depend on the concrete Dispatcher type.
type EventDispatcherPort interface {
	Dispatch(event platformevents.Event)
}

// InvoiceChangedEventPublisher adapts the platform Dispatcher to the
// InvoiceChangedPublisher port. It is constructed in bootstrap and injected
// wherever the invoicing service needs to publish the event.
type InvoiceChangedEventPublisher struct {
	dispatcher EventDispatcherPort
}

// NewInvoiceChangedEventPublisher returns a publisher backed by the given dispatcher.
func NewInvoiceChangedEventPublisher(dispatcher EventDispatcherPort) *InvoiceChangedEventPublisher {
	return &InvoiceChangedEventPublisher{dispatcher: dispatcher}
}

// PublishInvoiceChanged fires the invoice_changed event for invoiceID.
func (p *InvoiceChangedEventPublisher) PublishInvoiceChanged(_ context.Context, invoiceID string) error {
	if invoiceID == "" {
		return fmt.Errorf("invoice_changed publisher: invoiceID must not be empty")
	}
	p.dispatcher.Dispatch(newInvoiceChangedEvent(invoiceID))
	return nil
}
