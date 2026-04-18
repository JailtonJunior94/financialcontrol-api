package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type BillRequest struct {
	Date time.Time `json:"date"`
}

func (b *BillRequest) IsValid() error {
	if b.Date == time.Now() {
		return customerrors.DateIsRequired
	}

	return nil
}
