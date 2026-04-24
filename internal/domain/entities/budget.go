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
	zero, _ := vos.NewMoney(0, vos.CurrencyBRL)
	b.AmountUsed = zero
	for _, item := range b.Items {
		b.AmountUsed, _ = b.AmountUsed.Add(item.AmountUsed)
	}
}

func (b *Budget) CalculatePercentageUsed() {
	b.PercentageUsed = vos.Percentage{}
	for _, item := range b.Items {
		b.PercentageUsed, _ = b.PercentageUsed.Add(item.PercentageUsed)
	}
}

func (b *Budget) CalculatePercentageTotal() bool {
	var total vos.Percentage
	for _, item := range b.Items {
		total, _ = total.Add(item.PercentageGoal)
	}
	hundredPercent, _ := vos.NewPercentage(100000) // 100.000%
	return total.Equals(hundredPercent)
}

func NewBudgetItem(budget *Budget, category string, percentageGoal vos.Percentage) *BudgetItem {
	zero, _ := vos.NewMoney(0, vos.CurrencyBRL)
	zeroP, _ := vos.NewPercentage(0)
	hundredP, _ := vos.NewPercentage(100000)

	budgetItem := &BudgetItem{
		Budget:          budget,
		Category:        category,
		PercentageGoal:  percentageGoal,
		AmountUsed:      zero,
		PercentageUsed:  zeroP,
		PercentageTotal: hundredP,
	}

	budgetItem.CalculateAmountGoal()
	return budgetItem
}

func (b *BudgetItem) CalculateAmountGoal() {
	b.AmountGoal, _ = b.PercentageGoal.Apply(b.Budget.AmountGoal)
}

func (b *BudgetItem) AddAmountUsed(amount vos.Money) {
	b.AmountUsed, _ = b.AmountUsed.Add(amount)
	b.PercentageUsed, _ = b.PercentageUsed.Add(b.PercentageGoal)

	goalCents := b.Budget.AmountGoal.Cents()
	if goalCents > 0 {
		ratioPercent := float64(b.AmountUsed.Cents()) / float64(goalCents) * 100.0
		b.PercentageTotal, _ = vos.NewPercentageFromFloat(ratioPercent)
	}
}
