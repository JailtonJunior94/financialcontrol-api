package vos

import (
	"errors"
	"time"
)

type BillingCycle struct {
	closing ClosingDay
	due     DueDay
}

var ErrInvalidBillingCycle = errors.New("Ciclo de faturamento inválido")

func NewBillingCycle(closing ClosingDay, due DueDay) (BillingCycle, error) {
	if int(closing) == int(due) {
		return BillingCycle{}, ErrInvalidBillingCycle
	}
	return BillingCycle{closing: closing, due: due}, nil
}

// RehydrateBillingCycle restores persisted state without enforcing the
// creation-time invariant closing != due. This preserves read compatibility
// with legacy rows that predate the new validation rule.
func RehydrateBillingCycle(closing ClosingDay, due DueDay) BillingCycle {
	return BillingCycle{closing: closing, due: due}
}

// DueDateFor returns the due date for a purchase made on the given date (RF-04).
// When due < closing, the payment falls in the month after closing, so a base
// advance of 1 is added; purchases on or after the closing day advance one
// more month into the next cycle.
func (b BillingCycle) DueDateFor(purchase time.Time) time.Time {
	year, month := purchase.Year(), int(purchase.Month())

	advance := 0
	if int(b.due) < int(b.closing) {
		advance = 1
	}
	if purchase.Day() >= int(b.closing) {
		advance++
	}

	month += advance
	for month > 12 {
		month -= 12
		year++
	}

	dueMonth := time.Month(month)
	last := lastDayOfMonth(year, dueMonth)
	day := int(b.due)
	if day > last {
		day = last
	}
	return time.Date(year, dueMonth, day, 0, 0, 0, 0, purchase.Location())
}

// BestPurchaseDay returns the last day to make a purchase and still be included
// in the current billing cycle (closing - 1, minimum 1).
func (b BillingCycle) BestPurchaseDay() int {
	best := int(b.closing) - 1
	if best < 1 {
		return 1
	}
	return best
}

func (b BillingCycle) Closing() ClosingDay { return b.closing }
func (b BillingCycle) Due() DueDay         { return b.due }

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
