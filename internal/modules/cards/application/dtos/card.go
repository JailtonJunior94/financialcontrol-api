package dtos

import (
	"errors"
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
)

type CardRequest struct {
	FlagID         string    `json:"flag_id"`
	Name           string    `json:"name"`
	Number         string    `json:"number"`
	Description    string    `json:"description"`
	ClosingDay     int       `json:"closing_day"`
	ExpirationDate time.Time `json:"expiration_date"`
}

func (r CardRequest) Validate() error {
	var errs []error

	if _, err := vos.ParseFlagID(r.FlagID); err != nil {
		errs = append(errs, domain.ErrInvalidFlagID)
	}

	if _, err := vos.NewCardName(r.Name); err != nil {
		errs = append(errs, domain.ErrInvalidCardName)
	}

	if _, err := vos.NewCardNumber(r.Number); err != nil {
		errs = append(errs, domain.ErrInvalidCardNumber)
	}

	if _, err := vos.NewClosingDay(r.ClosingDay); err != nil {
		errs = append(errs, domain.ErrInvalidClosingDay)
	}

	if r.ExpirationDate.IsZero() {
		errs = append(errs, domain.ErrInvalidDueDay)
	}

	if !r.ExpirationDate.IsZero() {
		if _, err := vos.NewDueDay(r.ExpirationDate.Day()); err != nil {
			errs = append(errs, domain.ErrInvalidDueDay)
		}
	}

	return errors.Join(errs...)
}

type CardResponse struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Number         string       `json:"number,omitempty"`
	Description    string       `json:"description,omitempty"`
	ClosingDay     int          `json:"closing_day,omitempty"`
	BestDayToBuy   int          `json:"best_day_to_buy,omitempty"`
	ExpirationDate time.Time    `json:"expiration_date"`
	Active         bool         `json:"active"`
	Flag           FlagResponse `json:"flag"`
}

func ToCardResponse(c *entities.Card) CardResponse {
	resp := CardResponse{
		ID:             c.ID().String(),
		Name:           string(c.Name()),
		Number:         string(c.Number()),
		Description:    c.Description(),
		ClosingDay:     c.ClosingDay().Int(),
		BestDayToBuy:   c.BestPurchaseDay(),
		ExpirationDate: c.ExpirationDate(),
		Active:         c.Active(),
	}

	if f := c.Flag(); f != nil {
		resp.Flag = ToFlagResponse(*f)
	}

	return resp
}

func ToCardResponses(cards []entities.Card) []CardResponse {
	result := make([]CardResponse, len(cards))
	for i := range cards {
		result[i] = ToCardResponse(&cards[i])
	}
	return result
}
