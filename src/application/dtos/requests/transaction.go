package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type TransactionRequest struct {
	Date time.Time `json:"date"`
}

func (u *TransactionRequest) IsValid() error {
	if u.Date == time.Now() {
		return customErrors.DateIsRequired
	}

	return nil
}
