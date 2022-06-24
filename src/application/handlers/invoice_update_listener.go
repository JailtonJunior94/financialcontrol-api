package handlers

import (
	"errors"
	"log"
	"strings"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type invoiceChangedListener struct {
	data                  interface{}
	InvoiceRepository     interfaces.IInvoiceRepository
	TransactionRepository interfaces.ITransactionRepository
	TransactionService    usecases.ITransactionService
}

func NewInvoiceChangedListener(i interfaces.IInvoiceRepository,
	t usecases.ITransactionService,
	tr interfaces.ITransactionRepository,
) *invoiceChangedListener {
	return &invoiceChangedListener{
		InvoiceRepository:     i,
		TransactionService:    t,
		TransactionRepository: tr,
	}
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

	response := l.TransactionService.TransactionById(transaction.ID, transaction.UserId)
	transactionResponse := response.Data.(*responses.TransactionResponse)

	for _, i := range transactionResponse.Items {
		if strings.Contains(i.Title, invoice.Card.Description) {
			r := requests.NewTransactionItemRequest(invoice.Card.Description, "OUTCOME", invoice.Total)
			l.TransactionService.UpdateTransactionItem(transaction.ID, i.ID, transaction.UserId, r)
		}
	}

	return nil
}
