package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type DefaultCardService struct {
	repository CardRepository
}

func NewCardService(repository CardRepository) CardService {
	return &DefaultCardService{repository: repository}
}

func (s *DefaultCardService) Cards(userID string) *responses.HttpResponse {
	cards, err := s.repository.GetCards(userID)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(ToManyCardResponse(cards))
}

func (s *DefaultCardService) CardById(id, userID string) *responses.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	return responses.Ok(ToCardResponse(card))
}

func (s *DefaultCardService) CreateCard(userID string, request *CardRequest) *responses.HttpResponse {
	card, err := s.repository.AddCard(ToCardEntity(request, userID))
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(ToCardResponse(card))
}

func (s *DefaultCardService) UpdateCard(id, userID string, request *CardRequest) *responses.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	card.Update(request.FlagID, request.Name, request.Description, request.Number, request.ClosingDay, request.ExpirationDate)
	if _, err := s.repository.UpdateCard(card); err != nil {
		return responses.ServerError()
	}

	return responses.Ok(ToCardResponse(card))
}

func (s *DefaultCardService) RemoveCard(id, userID string) *responses.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	card.UpdateStatus(false)
	if _, err := s.repository.UpdateCard(card); err != nil {
		return responses.ServerError()
	}

	return responses.NoContent()
}
