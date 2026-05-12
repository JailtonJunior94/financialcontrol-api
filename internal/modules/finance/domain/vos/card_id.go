package vos

import (
	"errors"

	"github.com/google/uuid"
)

// CardID is a UUID-based value object identifying a card (cross-module reference).
type CardID string

var ErrInvalidCardID = errors.New("card ID inválido")

// NewCardID generates a new random CardID.
func NewCardID() CardID {
	return CardID(uuid.NewString())
}

// ParseCardID parses and validates a string as a CardID.
func ParseCardID(s string) (CardID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrInvalidCardID
	}
	return CardID(s), nil
}

func (c CardID) String() string { return string(c) }
