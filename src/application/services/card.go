package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type CardService struct {
	CardRepository interfaces.ICardRepository
}

func NewCardService(r interfaces.ICardRepository) usecases.ICardService {
	return &CardService{CardRepository: r}
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
	return nil
}
