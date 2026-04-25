package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
)

type CardRequest struct {
	FlagID         string    `json:"flagId"`
	Name           string    `json:"name"`
	Number         string    `json:"number"`
	Description    string    `json:"description"`
	ClosingDay     int       `json:"closingDay"`
	ExpirationDate time.Time `json:"expirationDate"`
}

func (c *CardRequest) IsValid() error {
	if c.Name == "" {
		return customerrors.NameIsRequired
	}

	return nil
}
