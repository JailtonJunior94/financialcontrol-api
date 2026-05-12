package vos

import (
	"errors"

	"github.com/google/uuid"
)

// TransactionID is a UUID-based value object identifying a transaction.
type TransactionID string

var ErrInvalidTransactionID = errors.New("transaction ID inválido")

// NewTransactionID generates a new random TransactionID.
func NewTransactionID() TransactionID {
	return TransactionID(uuid.NewString())
}

// ParseTransactionID parses and validates a string as a TransactionID.
func ParseTransactionID(s string) (TransactionID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrInvalidTransactionID
	}
	return TransactionID(s), nil
}

func (t TransactionID) String() string { return string(t) }
