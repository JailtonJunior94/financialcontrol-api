package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type ListCards interface {
	Execute(ctx context.Context, userID identityvo.UserID, pagination dtos.Pagination) ([]dtos.CardResponse, error)
}

type GetCard interface {
	Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID) (dtos.CardResponse, error)
}

type CreateCard interface {
	Execute(ctx context.Context, userID identityvo.UserID, req dtos.CardRequest) (dtos.CardResponse, error)
}

type UpdateCard interface {
	Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID, req dtos.CardRequest) (dtos.CardResponse, error)
}

type DeactivateCard interface {
	Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID) error
}

type ListFlags interface {
	Execute(ctx context.Context) ([]dtos.FlagResponse, error)
}
