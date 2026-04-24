package usecase

import (
	"fmt"
	"sync"

	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
)

type UpdateTransactionUseCase struct {
	TransactionRepository transactionsapp.TransactionRepository
	InvoiceRepository     invoicingapp.InvoiceRepository
	TransactionService    transactionsapp.TransactionAppService
}

func NewUpdateTransactionUseCase(
	r transactionsapp.TransactionRepository,
	i invoicingapp.InvoiceRepository,
	tt transactionsapp.TransactionAppService,
) *UpdateTransactionUseCase {
	return &UpdateTransactionUseCase{
		TransactionRepository: r,
		InvoiceRepository:     i,
		TransactionService:    tt,
	}
}

func (u *UpdateTransactionUseCase) Execute(wg *sync.WaitGroup, cardID string) error {
	defer wg.Done()

	invoices, err := u.InvoiceRepository.FetchInvoiceByCard(cardID)
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		transaction, err := u.TransactionRepository.FetchTransactionByDate(invoice.Date, invoice.Description)
		if err != nil {
			fmt.Println(err)
			continue
		}

		r := transactionsapp.NewTransactionItemRequest(invoice.Description, "OUTCOME", invoice.Total)
		if transaction == nil {
			continue
		}

		res := u.TransactionService.UpdateTransactionItem(transaction.TransactionID, transaction.ID, transaction.UserID, r)
		fmt.Printf("[StatusCode] [%d] [Message] [%v]\n", res.StatusCode, res.Data)
	}
	return nil
}
