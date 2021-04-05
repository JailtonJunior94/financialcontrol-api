package requests

import "github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"

type BillItemRequest struct {
	Title string  `json:"title"`
	Value float64 `json:"Value"`
}

func (b *BillItemRequest) IsValid() error {
	if b.Title == "" {
		return customErrors.TitleIsRequired
	}

	if b.Value == 0 {
		return customErrors.ValueIsRequired
	}

	return nil
}
