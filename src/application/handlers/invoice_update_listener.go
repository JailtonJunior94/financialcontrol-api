package handlers

import (
	"errors"
	"log"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
)

type invoiceChangedListener struct {
	data                  interface{}
	InvoiceRepository     interfaces.IInvoiceRepository
	TransactionRepository interfaces.ITransactionRepository
}

func NewInvoiceChangedListener(i interfaces.IInvoiceRepository, t interfaces.ITransactionRepository) *invoiceChangedListener {
	return &invoiceChangedListener{InvoiceRepository: i, TransactionRepository: t}
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

	if err = l.updateTransactionValue(invoice); err != nil {
		return err
	}

	log.Println("[Success] [Evento] [invoice_changed] [Processado com sucesso]")
	return nil
}

func (l *invoiceChangedListener) updateTransactionValue(invoice *entities.Invoice) error {
	if !invoice.MarkImportTransactions {
		log.Println("[Success] [Fatura não marcada para importação nas transações]")
		return nil
	}

	transaction, err := l.TransactionRepository.GetTransactionByDate(invoice.Date, invoice.Date, invoice.Card.UserId)
	if err != nil {
		return err
	}

	log.Println(transaction)
	return nil
}
