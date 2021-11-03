package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

func ToCardEntity(r *requests.CardRequest, userId string) (e *entities.Card) {
	card := entities.NewCard(userId, r.FlagId, r.Name, r.Description, r.Number, r.ClosingDay, r.ExpirationDate)
	return card
}

func ToCardResponse(e *entities.Card) (r *responses.CardResponse) {
	return &responses.CardResponse{
		ID:             e.ID,
		Name:           e.Name,
		Number:         e.Number,
		Description:    e.Description,
		ClosingDay:     e.ClosingDay,
		ExpirationDate: e.ExpirationDate,
		BestDayToBuy:   e.BestDayToBuy(),
		Active:         e.Active,
		Flag: responses.FlagResponse{
			ID:     e.Flag.ID,
			Name:   e.Flag.Name,
			Active: e.Flag.Active,
		},
	}
}

func ToManyCardResponse(entities []entities.Card) (r []responses.CardResponse) {
	if len(entities) == 0 {
		return make([]responses.CardResponse, 0)
	}

	for _, e := range entities {
		card := responses.CardResponse{
			ID:             e.ID,
			Name:           e.Name,
			Number:         e.Number,
			Description:    e.Description,
			ClosingDay:     e.ClosingDay,
			ExpirationDate: e.ExpirationDate,
			BestDayToBuy:   e.BestDayToBuy(),
			Active:         e.Active,
			Flag: responses.FlagResponse{
				ID:     e.Flag.ID,
				Name:   e.Flag.Name,
				Active: e.Flag.Active,
			},
		}
		r = append(r, card)
	}

	return r
}
