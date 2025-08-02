package usecase

import (
	"fmt"
	"sync"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type UpdateTransactionUseCase struct {
	TransactionRepository interfaces.ITransactionRepository
	InvoiceRepository     interfaces.IInvoiceRepository
	TransactionService    usecases.ITransactionService
}

func NewUpdateTransactionUseCase(r interfaces.ITransactionRepository,
	i interfaces.IInvoiceRepository,
	tt usecases.ITransactionService,
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

		r := requests.NewTransactionItemRequest(invoice.Description, "OUTCOME", invoice.Total)
		if transaction == nil {
			// t, _ := u.TransactionRepository.GetTransactionByDate(invoice.Date, invoice.Date, "F978F969-3EB6-4D0E-8E4E-3270A20F3513")

			// res := u.TransactionService.CreateTransactionItem(r, t.ID, t.UserId)
			// fmt.Printf("[StatusCode] [%d] [Message] [%v]\n", res.StatusCode, res.Data)
			continue
		}

		res := u.TransactionService.UpdateTransactionItem(transaction.TransactionID, transaction.ID, transaction.UserID, r)
		fmt.Printf("[StatusCode] [%d] [Message] [%v]\n", res.StatusCode, res.Data)
	}
	return nil
}
