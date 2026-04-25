package domain

import (
	"time"

	platformevents "github.com/jailtonjunior94/financialcontrol-api/pkg/events"
)

// InvoiceChangedPayload carries all data for the invoice_changed event.
// All fields are primitive or stdlib types — no references to domain entities.
// This ensures the payload is serializable and cross-module safe (ADR-003).
type InvoiceChangedPayload struct {
	InvoiceID              string
	CardDescription        string
	UserID                 string
	ReferenceDate          time.Time
	Total                  float64
	MarkImportTransactions bool
}

// invoiceChangedEvent is the domain event fired when an invoice total is
// recalculated. It is internal to the invoicing module.
type invoiceChangedEvent struct {
	payload InvoiceChangedPayload
}

// NewInvoiceChangedEvent constructs the event with the given payload.
func NewInvoiceChangedEvent(payload InvoiceChangedPayload) platformevents.Event {
	return &invoiceChangedEvent{payload: payload}
}

func (e *invoiceChangedEvent) GetKey() string {
	return "invoice_changed"
}

func (e *invoiceChangedEvent) GetData() any {
	return e.payload
}
