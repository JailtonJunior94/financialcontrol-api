package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	Budget struct {
		Date           time.Time
		AmountGoal     vos.Money
		AmountUsed     vos.Money
		PercentageUsed vos.Percentage
		Items          []*BudgetItem
	}

	BudgetItem struct {
		Budget          *Budget
		Category        string
		PercentageGoal  vos.Percentage
		AmountGoal      vos.Money
		AmountUsed      vos.Money
		PercentageUsed  vos.Percentage
		PercentageTotal vos.Percentage
	}
)

func NewBudget(date time.Time, amountGoal vos.Money) *Budget {
	return &Budget{
		Date:       date,
		AmountGoal: amountGoal,
	}
}

func (b *Budget) AddItems(items []*BudgetItem) bool {
	b.Items = append(b.Items, items...)
	b.CalculateAmountUsed()
	b.CalculatePercentageUsed()
	return b.CalculatePercentageTotal()
}

func (b *Budget) CalculateAmountUsed() {
	for _, item := range b.Items {
		b.AmountUsed = b.AmountUsed.Add(item.AmountUsed)
	}
}

func (b *Budget) CalculatePercentageUsed() {
	for _, item := range b.Items {
		b.PercentageUsed = b.PercentageUsed.Add(item.PercentageUsed)
	}
}

func (b *Budget) CalculatePercentageTotal() bool {
	var total vos.Percentage
	for _, item := range b.Items {
		total = total.Add(item.PercentageGoal)
	}
	return total.Equals(vos.NewPercentage(100))
}

func NewBudgetItem(budget *Budget, category string, percentageGoal vos.Percentage) *BudgetItem {
	budgetItem := &BudgetItem{
		Budget:          budget,
		Category:        category,
		PercentageGoal:  percentageGoal,
		AmountUsed:      vos.NewMoney(0),
		PercentageUsed:  vos.NewPercentage(0),
		PercentageTotal: vos.NewPercentage(100),
	}

	budgetItem.CalculateAmountGoal()
	return budgetItem
}

func (b *BudgetItem) CalculateAmountGoal() {
	b.AmountGoal = b.Budget.AmountGoal.Mul(b.PercentageGoal.Percentage())
}

func (b *BudgetItem) AddAmountUsed(amount vos.Money) {
	b.AmountUsed = b.AmountUsed.Add(amount)
	b.PercentageUsed = b.PercentageUsed.Add(b.PercentageGoal)

	total, _ := b.AmountUsed.Div(b.Budget.AmountGoal.Money())
	b.PercentageTotal = vos.NewPercentage(total.Mul(100).Money())
}
