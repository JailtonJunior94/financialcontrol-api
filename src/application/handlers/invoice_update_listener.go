package handlers

import (
	"errors"
	"log"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
)

type invoiceChangedListener struct {
	data              interface{}
	InvoiceRepository interfaces.IInvoiceRepository
}

func NewInvoiceChangedListener(i interfaces.IInvoiceRepository) *invoiceChangedListener {
	return &invoiceChangedListener{InvoiceRepository: i}
}

func (l *invoiceChangedListener) SetData(data interface{}) {
	l.data = data
}

func (l *invoiceChangedListener) Handle() error {
	invoice, err := l.InvoiceRepository.GetInvoiceById(l.data.(string))
	if err != nil {
		return errors.New("Não foi possível obter invoice por ID")
	}

	invoice.UpdatingValues()
	_, err = l.InvoiceRepository.UpdateInvoice(invoice)
	if err != nil {
		return errors.New("Não foi possível obter invoice por ID")
	}

	log.Println("[Success] [Evento] [invoice_changed] [Processado com sucesso] ")
	return nil
}
