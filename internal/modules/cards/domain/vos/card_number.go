package vos

import (
	"errors"
	"unicode"
)

type CardNumber string

var ErrInvalidCardNumber = errors.New("Número do cartão inválido")

func NewCardNumber(s string) (CardNumber, error) {
	if !isNumericOnly(s) || len(s) < 12 || len(s) > 19 {
		return "", ErrInvalidCardNumber
	}
	return CardNumber(s), nil
}

func isNumericOnly(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (n CardNumber) String() string { return string(n) }
