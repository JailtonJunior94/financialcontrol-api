package handlers

import "fmt"

type invoiceChangedListener struct {
	data interface{}
}

func NewInvoiceChangedListener() *invoiceChangedListener {
	return &invoiceChangedListener{}
}

func (l *invoiceChangedListener) SetData(data interface{}) {
	l.data = data
}

func (l *invoiceChangedListener) Handle() error {
	fmt.Printf("ID da fatura alterada: %s\n", l.data.(string))
	return nil
}
