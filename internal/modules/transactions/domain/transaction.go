package domain

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

type Transaction struct {
	UserId  string    `db:"UserId"`
	Date    time.Time `db:"Date"`
	Total   float64   `db:"Total"`
	Income  float64   `db:"Income"`
	Outcome float64   `db:"Outcome"`
	Entity
	TransactionItems []TransactionItem
}

func NewTransaction(userID string, date time.Time) *Transaction {
	transaction := &Transaction{
		UserId: userID,
		Date:   shared.NewTime(shared.Time{Date: date}).FormatDate(),
	}
	transaction.NewEntity()

	return transaction
}

func NewTransactionWithValues(date time.Time, userID string, total, income, outcome float64) *Transaction {
	transaction := &Transaction{
		Date:    shared.NewTime(shared.Time{Date: date}).FormatDate(),
		UserId:  userID,
		Total:   total,
		Income:  income,
		Outcome: outcome,
	}
	transaction.NewEntity()

	return transaction
}

func (t *Transaction) AddItems(items []TransactionItem) {
	t.TransactionItems = items
}

func (t *Transaction) UpdatingValues() {
	t.GetTotal()
	t.ChangeUpdatedAt()
}

func (t *Transaction) GetTotal() float64 {
	t.Total = t.SumIncomes() - t.SumOutcome()
	return t.Total
}

func (t *Transaction) SumIncomes() float64 {
	var income float64
	for _, item := range t.TransactionItems {
		if item.Type != Income {
			continue
		}

		income += item.Value
	}

	return t.AddIncome(income)
}

func (t *Transaction) AddIncome(income float64) float64 {
	t.Income = income
	return t.Income
}

func (t *Transaction) SumOutcome() float64 {
	var outcome float64
	for _, item := range t.TransactionItems {
		if item.Type != Outcome {
			continue
		}

		outcome += item.Value
	}

	return t.AddOutcome(outcome)
}

func (t *Transaction) AddOutcome(outcome float64) float64 {
	t.Outcome = outcome
	return t.Outcome
}
