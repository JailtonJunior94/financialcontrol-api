package vos

import (
	"errors"
	"net/mail"
)

type Email string

var ErrInvalidEmail = errors.New("e-mail inválido")

func NewEmail(value string) (Email, error) {
	if _, err := mail.ParseAddress(value); err != nil {
		return "", ErrInvalidEmail
	}
	return Email(value), nil
}

func (e Email) String() string { return string(e) }
