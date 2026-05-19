package ports

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
)

type FlagRepository interface {
	List(ctx context.Context) ([]entities.Flag, error)
	Exists(ctx context.Context, id vos.FlagID) (bool, error)
}
