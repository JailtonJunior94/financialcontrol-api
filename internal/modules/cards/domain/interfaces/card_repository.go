package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type CardRepository interface {
	List(ctx context.Context, userID identityvo.UserID, pagination vos.Pagination) ([]entities.Card, error)
	// GetByID returns the active card with the given id that belongs to userID.
	// It returns (nil, domain.ErrCardNotFound) when the card does not exist or
	// does not belong to the user. It never returns (nil, nil).
	GetByID(ctx context.Context, userID identityvo.UserID, id vos.CardID) (*entities.Card, error)
	Add(ctx context.Context, card *entities.Card) error
	Update(ctx context.Context, card *entities.Card) error
}
