package vos

import "errors"

type HashedPassword string

var ErrInvalidPassword = errors.New("senha inválida")

func NewHashedPassword(value string) (HashedPassword, error) {
	if value == "" {
		return "", ErrInvalidPassword
	}
	return HashedPassword(value), nil
}

func (h HashedPassword) String() string { return string(h) }
