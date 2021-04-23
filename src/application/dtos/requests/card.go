package requests

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type CardRequest struct {
	FlagId         string    `json:"flagId"`
	Name           string    `json:"name"`
	Number         string    `json:"number"`
	Description    string    `json:"description"`
	ClosingDay     int       `json:"closingDay"`
	ExpirationDate time.Time `json:"expirationDate"`
}

func (c *CardRequest) IsValid() error {
	if c.Name == "" {
		return customErrors.NameIsRequired
	}

	return nil
}
