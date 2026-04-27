package vos

import (
	"errors"

	"github.com/google/uuid"
)

type UserID string

var ErrInvalidUserID = errors.New("user ID inválido")

func NewUserID() UserID {
	return UserID(uuid.NewString())
}

func ParseUserID(s string) (UserID, error) {
	if s == "" {
		return "", ErrInvalidUserID
	}
	return UserID(s), nil
}

func (u UserID) String() string { return string(u) }
