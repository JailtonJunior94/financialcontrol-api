package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type BillRequest struct {
	Date time.Time `json:"date"`
}

func (b *BillRequest) IsValid() error {
	if b.Date == time.Now() {
		return customErrors.DateIsRequired
	}

	return nil
}
