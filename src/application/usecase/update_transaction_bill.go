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

func (u *UpdateTransactionBill) Execute() error {
	// transactions, _ := u.TransactionRepository.GetTransactions("F978F969-3EB6-4D0E-8E4E-3270A20F3513")
	// transactionsGroup := make(map[int][]*Transaction)

	// for _, t := range transactions {
	// 	if v, exists := transactionsGroup[t.Date.Year()]; exists {
	// 		transactionsGroup[t.Date.Year()] = append(v, NewTransaction(t.Date.Year(), t.Total, t.Income, t.Outcome))
	// 		continue
	// 	}
	// 	transactionsGroup[t.Date.Year()] = append(transactionsGroup[t.Date.Year()], NewTransaction(t.Date.Year(), t.Total, t.Income, t.Outcome))
	// }

	// for _, t := range transactionsGroup {
	// 	for _, tt := range t {
	// 		fmt.Printf("ANO: %d | Porcentagem de despesa %v%%\n", tt.Year, math.Round((tt.Outcome/tt.Income)*100))
	// 	}
	// }

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

		r := requests.NewTransactionItemRequest("Casa (Despesas)", "OUTCOME", value)
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
