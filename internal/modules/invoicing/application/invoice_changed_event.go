package invoicingapp

import platformevents "github.com/jailtonjunior94/financialcontrol-api/internal/platform/events"

// invoiceChangedEvent is the domain event fired when an invoice total is
// recalculated. It is internal to the invoicing module.
type invoiceChangedEvent struct {
	data string
}

func newInvoiceChangedEvent(invoiceID string) platformevents.Event {
	return &invoiceChangedEvent{data: invoiceID}
}

func (e *invoiceChangedEvent) GetKey() string {
	return "invoice_changed"
}

func (e *invoiceChangedEvent) GetData() interface{} {
	return e.data
}
