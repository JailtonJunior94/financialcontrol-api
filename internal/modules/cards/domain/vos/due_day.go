package vos

import "errors"

type DueDay int

//nolint:staticcheck // Preserva mensagens públicas de validação já expostas pelo módulo.
var ErrInvalidDueDay = errors.New("Dia de vencimento inválido (1..31)")

func NewDueDay(n int) (DueDay, error) {
	if n < 1 || n > 31 {
		return 0, ErrInvalidDueDay
	}
	return DueDay(n), nil
}

func MustDueDay(n int) DueDay {
	d, err := NewDueDay(n)
	if err != nil {
		panic(err)
	}
	return d
}

func (d DueDay) Int() int { return int(d) }
