package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CardRepository interface {
	List(ctx context.Context, userID identityvo.UserID) ([]entities.Card, error)
	GetByID(ctx context.Context, userID identityvo.UserID, id vos.CardID) (*entities.Card, error)
	Add(ctx context.Context, card *entities.Card) error
	Update(ctx context.Context, card *entities.Card) error
}
