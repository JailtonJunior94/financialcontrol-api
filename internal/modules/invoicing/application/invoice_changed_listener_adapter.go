package invoicingapp

import (
	"context"
	"log"
)

// InvoiceChangedListenerAdapter bridges the legacy domain event dispatcher
// (which uses SetData/Handle without context) to InvoiceChangedHandler.
// It is registered in bootstrap so the event dispatcher can call it when the
// invoice_changed event is fired.
type InvoiceChangedListenerAdapter struct {
	handler *InvoiceChangedHandler
	data    interface{}
}

// NewInvoiceChangedListenerAdapter wraps the modular handler for registration
// in the legacy event dispatcher.
func NewInvoiceChangedListenerAdapter(handler *InvoiceChangedHandler) *InvoiceChangedListenerAdapter {
	return &InvoiceChangedListenerAdapter{handler: handler}
}

// SetData stores the event payload (invoiceID string) for the next Handle call.
func (a *InvoiceChangedListenerAdapter) SetData(data interface{}) {
	a.data = data
}

// Handle processes the stored event payload using the modular handler.
func (a *InvoiceChangedListenerAdapter) Handle() error {
	invoiceID, ok := a.data.(string)
	if !ok {
		log.Printf("[invoicing] invoice_changed listener: unexpected data type %T", a.data)
		return nil
	}

	if err := a.handler.Handle(context.Background(), invoiceID); err != nil {
		log.Printf("[invoicing] invoice_changed listener: %v", err)
		return err
	}

	return nil
}
