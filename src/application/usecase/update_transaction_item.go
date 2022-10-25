package usecase

import (
	"fmt"

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

func (u *UpdateTransactionUseCase) Execute() error {
	invoices, err := u.InvoiceRepository.FetchInvoiceByCard("45DE5288-D5D0-471A-BF18-09FE1FD2FC86")
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		transaction, err := u.TransactionRepository.FetchTransactionByDate(invoice.Date, invoice.Description)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if transaction == nil {
			fmt.Println(err)
			continue
		}

		r := requests.NewTransactionItemRequest(invoice.Description, "OUTCOME", invoice.Total)
		res := u.TransactionService.UpdateTransactionItem(transaction.TransactionID, transaction.ID, transaction.UserID, r)
		fmt.Printf("[StatusCode] [%d] [Message] [%v]\n", res.StatusCode, res.Data)
	}
	return nil
}
