package interfaces

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
)

type ICardRepository interface {
	GetCards(userId string) (cards []entities.Card, err error)
	GetCardById(id string, userId string) (card *entities.Card, err error)
	AddCard(c *entities.Card) (card *entities.Card, err error)
	UpdateCard(c *entities.Card) (card *entities.Card, err error)
}
