package invoicingapp

import (
	"context"
	"log"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
)

// InvoiceChangedListenerAdapter bridges the platform event dispatcher
// (which uses SetData/Handle without context) to InvoiceChangedHandler.
// It is registered in bootstrap so the event dispatcher can call it when the
// invoice_changed event is fired.
type InvoiceChangedListenerAdapter struct {
	handler *InvoiceChangedHandler
	data    any
}

// NewInvoiceChangedListenerAdapter wraps the modular handler for registration
// in the event dispatcher.
func NewInvoiceChangedListenerAdapter(handler *InvoiceChangedHandler) *InvoiceChangedListenerAdapter {
	return &InvoiceChangedListenerAdapter{handler: handler}
}

// SetData stores the event payload for the next Handle call.
func (a *InvoiceChangedListenerAdapter) SetData(data any) {
	a.data = data
}

// Handle processes the stored event payload using the modular handler.
func (a *InvoiceChangedListenerAdapter) Handle() error {
	payload, ok := a.data.(invoicingdomain.InvoiceChangedPayload)
	if !ok {
		log.Printf("[invoicing] invoice_changed listener: unexpected data type %T", a.data)
		return nil
	}

	if err := a.handler.Handle(context.Background(), payload); err != nil {
		log.Printf("[invoicing] invoice_changed listener: %v", err)
		return err
	}

	return nil
}
