package vos

import (
	"errors"

	"github.com/google/uuid"
)

// InstallmentID is a UUID-based value object identifying an installment.
type InstallmentID string

var ErrInvalidInstallmentID = errors.New("installment ID inválido")

// NewInstallmentID generates a new random InstallmentID.
func NewInstallmentID() InstallmentID {
	return InstallmentID(uuid.NewString())
}

// ParseInstallmentID parses and validates a string as an InstallmentID.
func ParseInstallmentID(s string) (InstallmentID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrInvalidInstallmentID
	}
	return InstallmentID(s), nil
}

func (i InstallmentID) String() string { return string(i) }
