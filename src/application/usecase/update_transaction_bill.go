package usecase

import (
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type UpdateTransactionBill struct {
	BillRepository        interfaces.IBillRepository
	TransactionService    usecases.ITransactionService
	TransactionRepository interfaces.ITransactionRepository
}

func NewUpdateTransactionBill(
	b interfaces.IBillRepository,
	ts usecases.ITransactionService,
	t interfaces.ITransactionRepository,
) *UpdateTransactionBill {
	return &UpdateTransactionBill{
		BillRepository:        b,
		TransactionService:    ts,
		TransactionRepository: t,
	}
}

func (u *UpdateTransactionBill) Execute() error {
	bills, err := u.BillRepository.GetBills()
	if err != nil {
		return err
	}

	for _, bill := range bills {
		transaction, err := u.TransactionRepository.FetchTransactionByDate(bill.Date, "Casa (Despesas)")
		if err != nil {
			fmt.Println(err)
			continue
		}

		r := requests.NewTransactionItemRequest("Casa (Despesas)", "OUTCOME", bill.SixtyPercent)
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
