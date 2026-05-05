package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type updateCard struct {
	cardRepo interfaces.CardRepository
	flagRepo interfaces.FlagRepository
}

func NewUpdateCard(cardRepo interfaces.CardRepository, flagRepo interfaces.FlagRepository) UpdateCard {
	return &updateCard{cardRepo: cardRepo, flagRepo: flagRepo}
}

func (uc *updateCard) Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID, req dtos.CardRequest) (dtos.CardResponse, error) {
	if err := req.Validate(); err != nil {
		return dtos.CardResponse{}, err
	}

	card, err := uc.cardRepo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase update_card: %w", err)
	}

	flagID, err := vos.ParseFlagID(req.FlagID)
	if err != nil {
		return dtos.CardResponse{}, err
	}

	if err := card.Update(flagID, req.Name, req.Number, req.Description, req.ClosingDay, req.ExpirationDate); err != nil {
		return dtos.CardResponse{}, err
	}

	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase update_card: %w", err)
	}

	reloaded, err := uc.cardRepo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase update_card: %w", err)
	}

	return dtos.ToCardResponse(reloaded), nil
}
