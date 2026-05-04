package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type DefaultCardService struct {
	repository CardRepository
}

func NewCardService(repository CardRepository) CardService {
	return &DefaultCardService{repository: repository}
}

func (s *DefaultCardService) Cards(userID string) *web.HttpResponse {
	cards, err := s.repository.GetCards(userID)
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyCardResponse(cards))
}

func (s *DefaultCardService) CardById(id, userID string) *web.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return web.ServerError()
	}

	if card == nil {
		return web.NotFound(domain.ErrCardNotFound)
	}

	return web.Ok(ToCardResponse(card))
}

func (s *DefaultCardService) CreateCard(userID string, request *CardRequest) *web.HttpResponse {
	card, err := s.repository.AddCard(ToCardEntity(request, userID))
	if err != nil {
		return web.ServerError()
	}

	return web.Created(ToCardResponse(card))
}

func (s *DefaultCardService) UpdateCard(id, userID string, request *CardRequest) *web.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return web.ServerError()
	}

	if card == nil {
		return web.NotFound(domain.ErrCardNotFound)
	}

	card.Update(request.FlagID, request.Name, request.Description, request.Number, request.ClosingDay, request.ExpirationDate)
	if _, err := s.repository.UpdateCard(card); err != nil {
		return web.ServerError()
	}

	return web.Ok(ToCardResponse(card))
}

func (s *DefaultCardService) RemoveCard(id, userID string) *web.HttpResponse {
	card, err := s.repository.GetCardById(id, userID)
	if err != nil {
		return web.ServerError()
	}

	if card == nil {
		return web.NotFound(domain.ErrCardNotFound)
	}

	card.UpdateStatus(false)
	if _, err := s.repository.UpdateCard(card); err != nil {
		return web.ServerError()
	}

	return web.NoContent()
}
