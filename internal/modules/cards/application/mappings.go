package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"

func ToCardEntity(request *CardRequest, userID string) *entities.Card {
	return entities.NewCard(
		userID,
		request.FlagID,
		request.Name,
		request.Description,
		request.Number,
		request.ClosingDay,
		request.ExpirationDate,
	)
}

func ToCardResponse(card *entities.Card) *CardResponse {
	return &CardResponse{
		ID:             card.ID,
		Name:           card.Name,
		Number:         card.Number,
		Description:    card.Description,
		ClosingDay:     card.ClosingDay,
		BestDayToBuy:   card.BestDayToBuy(),
		ExpirationDate: card.ExpirationDate,
		Active:         card.Active,
		Flag: FlagResponse{
			ID:     card.Flag.ID,
			Name:   card.Flag.Name,
			Active: card.Flag.Active,
		},
	}
}

func ToManyCardResponse(cards []entities.Card) []CardResponse {
	if len(cards) == 0 {
		return make([]CardResponse, 0)
	}

	result := make([]CardResponse, 0, len(cards))
	for _, card := range cards {
		result = append(result, *ToCardResponse(&card))
	}

	return result
}
