package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type CardRepository interface {
	GetCards(userID string) ([]domain.Card, error)
	GetCardById(id, userID string) (*domain.Card, error)
	AddCard(card *domain.Card) (*domain.Card, error)
	UpdateCard(card *domain.Card) (*domain.Card, error)
}

type CardService interface {
	Cards(userID string) *web.HttpResponse
	CardById(id, userID string) *web.HttpResponse
	CreateCard(userID string, request *CardRequest) *web.HttpResponse
	UpdateCard(id, userID string, request *CardRequest) *web.HttpResponse
	RemoveCard(id, userID string) *web.HttpResponse
}
