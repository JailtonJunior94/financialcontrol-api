package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
)

type CardRepository interface {
	GetCards(userID string) ([]entities.Card, error)
	GetCardById(id, userID string) (*entities.Card, error)
	AddCard(card *entities.Card) (*entities.Card, error)
	UpdateCard(card *entities.Card) (*entities.Card, error)
}

type CardService interface {
	Cards(userID string) *web.HttpResponse
	CardById(id, userID string) *web.HttpResponse
	CreateCard(userID string, request *CardRequest) *web.HttpResponse
	UpdateCard(id, userID string, request *CardRequest) *web.HttpResponse
	RemoveCard(id, userID string) *web.HttpResponse
}

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}
