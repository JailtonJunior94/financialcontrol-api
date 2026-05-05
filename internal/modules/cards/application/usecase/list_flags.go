package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
)

type listFlags struct {
	flagRepo interfaces.FlagRepository
}

func NewListFlags(flagRepo interfaces.FlagRepository) ListFlags {
	return &listFlags{flagRepo: flagRepo}
}

func (uc *listFlags) Execute(ctx context.Context) ([]dtos.FlagResponse, error) {
	flags, err := uc.flagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase list_flags: %w", err)
	}
	return dtos.ToFlagResponses(flags), nil
}
