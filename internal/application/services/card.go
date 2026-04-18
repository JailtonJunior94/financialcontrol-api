package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
)

type CardService struct {
	CardRepository interfaces.ICardRepository
}

func NewCardService(r interfaces.ICardRepository) usecases.ICardService {
	return &CardService{CardRepository: r}
}

func (s *CardService) Cards(userId string) *responses.HttpResponse {
	cards, err := s.CardRepository.GetCards(userId)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyCardResponse(cards))
}

func (s *CardService) CardById(id, userId string) *responses.HttpResponse {
	card, err := s.CardRepository.GetCardById(id, userId)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	return responses.Ok(mappings.ToCardResponse(card))
}

func (s *CardService) CreateCard(userId string, request *requests.CardRequest) *responses.HttpResponse {
	newCard := mappings.ToCardEntity(request, userId)
	card, err := s.CardRepository.AddCard(newCard)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(mappings.ToCardResponse(card))
}

func (s *CardService) UpdateCard(id, userId string, request *requests.CardRequest) *responses.HttpResponse {
	card, err := s.CardRepository.GetCardById(id, userId)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	card.Update(request.FlagId, request.Name, request.Description, request.Number, request.ClosingDay, request.ExpirationDate)
	_, err = s.CardRepository.UpdateCard(card)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToCardResponse(card))
}

func (s *CardService) RemoveCard(id, userId string) *responses.HttpResponse {
	card, err := s.CardRepository.GetCardById(id, userId)
	if err != nil {
		return responses.ServerError()
	}

	if card == nil {
		return responses.NotFound(customerrors.CardNotFound)
	}

	card.UpdateStatus(false)
	_, err = s.CardRepository.UpdateCard(card)
	if err != nil {
		return responses.ServerError()
	}

	return responses.NoContent()
}
