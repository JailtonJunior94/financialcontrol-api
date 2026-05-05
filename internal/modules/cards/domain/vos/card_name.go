package vos

import (
	"errors"
	"strings"
)

type CardName string

var ErrInvalidCardName = errors.New("Nome do cartão inválido")

func NewCardName(s string) (CardName, error) {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) < 1 || len(trimmed) > 100 {
		return "", ErrInvalidCardName
	}
	return CardName(trimmed), nil
}

func (n CardName) String() string { return string(n) }
