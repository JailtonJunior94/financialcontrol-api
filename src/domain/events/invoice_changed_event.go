package events

type invoiceChangedEvent struct {
	data string
}

func NewInvoiceChangedEvent(data interface{}) *invoiceChangedEvent {
	return &invoiceChangedEvent{
		data: data.(string),
	}
}

func (e *invoiceChangedEvent) GetKey() string {
	return "invoice_changed"
}

func (e *invoiceChangedEvent) GetData() interface{} {
	return e.data
}
