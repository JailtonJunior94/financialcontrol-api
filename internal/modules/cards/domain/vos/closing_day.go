package vos

import "errors"

type ClosingDay int

var ErrInvalidClosingDay = errors.New("Melhor dia de compra inválido (1..31)")

func NewClosingDay(n int) (ClosingDay, error) {
	if n < 1 || n > 31 {
		return 0, ErrInvalidClosingDay
	}
	return ClosingDay(n), nil
}

func MustClosingDay(n int) ClosingDay {
	c, err := NewClosingDay(n)
	if err != nil {
		panic(err)
	}
	return c
}

func (c ClosingDay) Int() int { return int(c) }
