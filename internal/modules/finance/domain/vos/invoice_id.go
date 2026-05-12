package vos

import (
	"errors"

	"github.com/google/uuid"
)

// InvoiceID is a UUID-based value object identifying an invoice.
type InvoiceID string

var ErrInvalidInvoiceID = errors.New("invoice ID inválido")

// NewInvoiceID generates a new random InvoiceID.
func NewInvoiceID() InvoiceID {
	return InvoiceID(uuid.NewString())
}

// ParseInvoiceID parses and validates a string as an InvoiceID.
func ParseInvoiceID(s string) (InvoiceID, error) {
	if _, err := uuid.Parse(s); err != nil {
		return "", ErrInvalidInvoiceID
	}
	return InvoiceID(s), nil
}

func (i InvoiceID) String() string { return string(i) }
