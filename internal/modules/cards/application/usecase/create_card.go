package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type createCard struct {
	cardRepo interfaces.CardRepository
	flagRepo interfaces.FlagRepository
}

func NewCreateCard(cardRepo interfaces.CardRepository, flagRepo interfaces.FlagRepository) CreateCard {
	return &createCard{cardRepo: cardRepo, flagRepo: flagRepo}
}

func (uc *createCard) Execute(ctx context.Context, userID identityvo.UserID, req dtos.CardRequest) (dtos.CardResponse, error) {
	if err := req.Validate(); err != nil {
		return dtos.CardResponse{}, err
	}

	flagID, err := vos.ParseFlagID(req.FlagID)
	if err != nil {
		return dtos.CardResponse{}, err
	}

	card, err := entities.NewCard(userID, flagID, req.Name, req.Number, req.Description, req.ClosingDay, req.ExpirationDate)
	if err != nil {
		return dtos.CardResponse{}, err
	}

	if err := uc.cardRepo.Add(ctx, card); err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase create_card: %w", err)
	}

	reloaded, err := uc.cardRepo.GetByID(ctx, userID, card.ID())
	if err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase create_card: %w", err)
	}

	return dtos.ToCardResponse(reloaded), nil
}
