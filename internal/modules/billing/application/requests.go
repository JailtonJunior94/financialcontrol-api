package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type BillRequest struct {
	Date time.Time `json:"date"`
}

func (b *BillRequest) IsValid() error {
	if b.Date.IsZero() {
		return customerrors.DateIsRequired
	}
	return nil
}

type BillItemRequest struct {
	Title string  `json:"title"`
	Value float64 `json:"Value"`
}

func (b *BillItemRequest) IsValid() error {
	if b.Title == "" {
		return customerrors.TitleIsRequired
	}
	if b.Value == 0 {
		return customerrors.ValueIsRequired
	}
	return nil
}
