package usecase

import (
	"fmt"
	"sync"

	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
)

type UpdateTransactionBill struct {
	BillRepository        billingapp.BillRepository
	TransactionService    transactionsapp.TransactionAppService
	TransactionRepository transactionsapp.TransactionRepository
}

func NewUpdateTransactionBill(
	b billingapp.BillRepository,
	ts transactionsapp.TransactionAppService,
	t transactionsapp.TransactionRepository,
) *UpdateTransactionBill {
	return &UpdateTransactionBill{
		BillRepository:        b,
		TransactionService:    ts,
		TransactionRepository: t,
	}
}

type Transaction struct {
	Year    int
	Total   float64
	Income  float64
	Outcome float64
}

func NewTransaction(year int, total, income, outcome float64) *Transaction {
	return &Transaction{
		Year:    year,
		Total:   total,
		Income:  income,
		Outcome: outcome,
	}
}

func (u *UpdateTransactionBill) Execute(wg *sync.WaitGroup) error {
	defer wg.Done()

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

		value := 0.0
		if bill.Date.Year() == 2023 {
			value = bill.Total / 2
		}

		if bill.Date.Year() != 2023 {
			value = bill.Total
		}

		r := transactionsapp.NewTransactionItemRequest("Casa (Despesas)", "OUTCOME", value)
		if transaction == nil {
			continue
		}

		res := u.TransactionService.UpdateTransactionItem(transaction.TransactionID, transaction.ID, transaction.UserID, r)
		fmt.Printf("[StatusCode] [%d] [Message] [%v]\n", res.StatusCode, res.Data)
	}
	return nil
}
