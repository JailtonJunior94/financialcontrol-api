package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type deactivateCard struct {
	cardRepo interfaces.CardRepository
}

func NewDeactivateCard(cardRepo interfaces.CardRepository) DeactivateCard {
	return &deactivateCard{cardRepo: cardRepo}
}

func (uc *deactivateCard) Execute(ctx context.Context, userID identityvo.UserID, id vos.CardID) error {
	card, err := uc.cardRepo.GetByID(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("usecase deactivate_card: %w", err)
	}

	// Idempotência: se já está inativo, não há nada a fazer e evitamos um UPDATE redundante.
	if !card.Active() {
		return nil
	}

	card.Deactivate()

	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return fmt.Errorf("usecase deactivate_card: %w", err)
	}

	return nil
}
