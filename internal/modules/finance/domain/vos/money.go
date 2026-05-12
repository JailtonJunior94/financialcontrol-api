package vos

import (
	"encoding/json"
	"fmt"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/shopspring/decimal"
)

const (
	currencyBRL      = "BRL"
	moneyPrecision   = 4
	displayPrecision = 2
)

// Money represents a monetary value with BRL currency.
// It is immutable: all arithmetic methods return a new Money value.
type Money struct {
	amount   decimal.Decimal
	currency string
}

// NewMoney constructs a Money from a string representation (e.g., "123.45").
// Only BRL currency is supported.
func NewMoney(raw string) (Money, error) {
	d, err := decimal.NewFromString(raw)
	if err != nil {
		return Money{}, domain.ErrInvalidMoneyFormat
	}
	return Money{amount: d, currency: currencyBRL}, nil
}

// NewMoneyFromDecimal constructs a Money directly from a decimal.Decimal.
func NewMoneyFromDecimal(d decimal.Decimal) Money {
	return Money{amount: d, currency: currencyBRL}
}

// Zero returns a zero-value Money in BRL.
func ZeroMoney() Money {
	return Money{amount: decimal.Zero, currency: currencyBRL}
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() string        { return m.currency }

// Add returns m + other. Panics if currencies differ (should not happen — BRL-only).
func (m Money) Add(other Money) Money {
	return Money{amount: m.amount.Add(other.amount), currency: m.currency}
}

// Sub returns m - other.
func (m Money) Sub(other Money) Money {
	return Money{amount: m.amount.Sub(other.amount), currency: m.currency}
}

// Mul returns m * n (integer multiplier).
func (m Money) Mul(n int) Money {
	return Money{amount: m.amount.Mul(decimal.NewFromInt(int64(n))), currency: m.currency}
}

// DivideEvenly splits m into n parts using half-even rounding.
// The remainder (due to integer cents) is added to the first part, so sum(parts) == m exactly.
// Precision is kept at 4 decimal places (DECIMAL(19,4)) — decisão G1.c.
func (m Money) DivideEvenly(n int) []Money {
	if n <= 0 {
		return nil
	}
	base := m.amount.DivRound(decimal.NewFromInt(int64(n)), moneyPrecision)
	parts := make([]Money, n)
	for i := 1; i < n; i++ {
		parts[i] = Money{amount: base, currency: m.currency}
	}
	// First part absorbs any rounding residue to guarantee exact sum.
	rest := base.Mul(decimal.NewFromInt(int64(n - 1)))
	parts[0] = Money{amount: m.amount.Sub(rest), currency: m.currency}
	return parts
}

// Equal reports whether m and other represent the same monetary value.
func (m Money) Equal(other Money) bool {
	return m.currency == other.currency && m.amount.Equal(other.amount)
}

// IsPositive reports whether m > 0.
func (m Money) IsPositive() bool { return m.amount.IsPositive() }

// IsNegative reports whether m < 0.
func (m Money) IsNegative() bool { return m.amount.IsNegative() }

// IsZero reports whether m == 0.
func (m Money) IsZero() bool { return m.amount.IsZero() }

// String returns a human-readable representation rounded to 2 decimal places.
func (m Money) String() string {
	return m.amount.StringFixed(displayPrecision)
}

// MarshalJSON serialises Money as a JSON string "123.45" — decisão H3.a.
func (m Money) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%s"`, m.amount.StringFixed(displayPrecision))
	return []byte(s), nil
}

// UnmarshalJSON deserialises Money from a JSON string.
// Rejects raw numbers to enforce precision — decisão H3.a.
func (m *Money) UnmarshalJSON(data []byte) error {
	// data must be a quoted string; if the first byte is not '"' it is a raw number.
	if len(data) == 0 || data[0] != '"' {
		return domain.ErrInvalidMoneyFormat
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return domain.ErrInvalidMoneyFormat
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return domain.ErrInvalidMoneyFormat
	}
	m.amount = d
	m.currency = currencyBRL
	return nil
}
