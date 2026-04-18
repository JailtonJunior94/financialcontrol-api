package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type TransactionRequest struct {
	Date time.Time `json:"date"`
}

func (u *TransactionRequest) IsValid() error {
	if u.Date == time.Now() {
		return customerrors.DateIsRequired
	}

	return nil
}
