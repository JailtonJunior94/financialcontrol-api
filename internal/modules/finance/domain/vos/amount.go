package vos

import domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"

// Amount is a Money value guaranteed to be strictly positive (> 0).
// It enforces RF-05: transaction amounts must be positive.
type Amount struct {
	money Money
}

// NewAmount constructs an Amount from a string. Returns ErrInvalidAmount when
// the parsed value is zero or negative.
func NewAmount(raw string) (Amount, error) {
	m, err := NewMoney(raw)
	if err != nil {
		return Amount{}, err
	}
	if !m.IsPositive() {
		return Amount{}, domain.ErrInvalidAmount
	}
	return Amount{money: m}, nil
}

// NewAmountFromMoney constructs an Amount from an existing Money value.
func NewAmountFromMoney(m Money) (Amount, error) {
	if !m.IsPositive() {
		return Amount{}, domain.ErrInvalidAmount
	}
	return Amount{money: m}, nil
}

// Money returns the underlying Money value.
func (a Amount) Money() Money { return a.money }

func (a Amount) String() string { return a.money.String() }
