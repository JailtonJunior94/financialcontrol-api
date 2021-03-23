package entities

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
)

type Transaction struct {
	UserId  string    `db:"UserId"`
	Date    time.Time `db:"Date"`
	Total   float64   `db:"Total"`
	Income  float64   `db:"Income"`
	Outcome float64   `db:"Outcome"`
	Entity
	User             User
	TransactionItems []TransactionItem
}

func (u *Transaction) NewTransaction(userId string, date time.Time) {
	u.Entity.NewEntity()
	u.UserId = userId
	u.Date = date
}

func (u *Transaction) AddItems(items []TransactionItem) {
	u.TransactionItems = items
}

func (u *Transaction) UpdatingValues() {
	u.GetTotal()
	u.ChangeUpdatedAt()
}

func (u *Transaction) GetTotal() float64 {
	u.Total = u.SumIncomes() - u.SumOutcome()
	return u.Total
}

func (u *Transaction) SumIncomes() float64 {
	filterByIncome := func(ti TransactionItem) bool {
		return ti.Type == constants.Income
	}

	incomes := filter(u.TransactionItems, filterByIncome)
	if len(incomes) == 0 {
		return u.AddIncome(0)
	}

	var income float64
	for _, i := range incomes {
		income += i.Value
	}

	return u.AddIncome(income)
}

func (u *Transaction) AddIncome(income float64) float64 {
	u.Income = income
	return u.Income
}

func (u *Transaction) SumOutcome() float64 {
	filterByOutcome := func(ti TransactionItem) bool {
		return ti.Type == constants.Outcome
	}

	outcomes := filter(u.TransactionItems, filterByOutcome)
	if len(outcomes) == 0 {
		return u.AddOutcome(0)
	}

	var outcome float64
	for _, o := range outcomes {
		outcome += o.Value
	}

	return u.AddOutcome(outcome)
}

func (u *Transaction) AddOutcome(outcome float64) float64 {
	u.Outcome = outcome
	return u.Outcome
}

func filter(ti []TransactionItem, cond func(TransactionItem) bool) (r []TransactionItem) {
	for _, i := range ti {
		if cond(i) {
			r = append(r, i)
		}
	}
	return
}
