package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
)

type CardRepository interface {
	GetCards(userID string) ([]entities.Card, error)
	GetCardById(id, userID string) (*entities.Card, error)
	AddCard(card *entities.Card) (*entities.Card, error)
	UpdateCard(card *entities.Card) (*entities.Card, error)
}

type CardService interface {
	Cards(userID string) *responses.HttpResponse
	CardById(id, userID string) *responses.HttpResponse
	CreateCard(userID string, request *CardRequest) *responses.HttpResponse
	UpdateCard(id, userID string, request *CardRequest) *responses.HttpResponse
	RemoveCard(id, userID string) *responses.HttpResponse
}

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}
