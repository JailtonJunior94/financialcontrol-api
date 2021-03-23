package requests

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type TransactionItemRequest struct {
	Title string  `json:"title"`
	Value float64 `json:"Value"`
	Type  string  `json:"Type"`
}

func (t *TransactionItemRequest) IsValid() error {
	if t.Title == "" {
		return customErrors.TitleIsRequired
	}

	if t.Value == 0 {
		return customErrors.ValueIsRequired
	}

	if t.Type == "" {
		return customErrors.TypeIsRequired
	}

	return nil
}
