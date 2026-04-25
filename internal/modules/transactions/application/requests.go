package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
)

type TransactionRequest struct {
	Date time.Time `json:"date"`
}

func (r *TransactionRequest) IsValid() error {
	if r.Date.IsZero() {
		return customerrors.DateIsRequired
	}

	return nil
}

type TransactionItemRequest struct {
	Title string  `json:"title"`
	Value float64 `json:"Value"`
	Type  string  `json:"Type"`
}

func (r *TransactionItemRequest) IsValid() error {
	if r.Title == "" {
		return customerrors.TitleIsRequired
	}

	if r.Value == 0 {
		return customerrors.ValueIsRequired
	}

	if r.Type == "" {
		return customerrors.TypeIsRequired
	}

	return nil
}

func NewTransactionItemRequest(title, transactionType string, value float64) *TransactionItemRequest {
	return &TransactionItemRequest{
		Title: title,
		Type:  transactionType,
		Value: value,
	}
}

type TransactionMarkAsPaid struct {
	MarkAsPaid bool `json:"markAsPaid"`
}
