package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type getCard struct {
	cardRepo interfaces.CardRepository
}

func NewGetCard(cardRepo interfaces.CardRepository) GetCard {
	return &getCard{cardRepo: cardRepo}
}

func (uc *getCard) Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID) (dtos.CardResponse, error) {
	card, err := uc.cardRepo.GetByID(ctx, userID, id)
	if err != nil {
		return dtos.CardResponse{}, fmt.Errorf("usecase get_card: %w", err)
	}
	return dtos.ToCardResponse(card), nil
}
