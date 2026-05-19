package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type listCards struct {
	cardRepo ports.CardRepository
}

func NewListCards(cardRepo ports.CardRepository) ListCards {
	return &listCards{cardRepo: cardRepo}
}

func (uc *listCards) Execute(ctx context.Context, userID identityvo.UserID, pagination dtos.Pagination) ([]dtos.CardResponse, error) {
	pagination = dtos.NewPagination(pagination.Page, pagination.Size)
	cards, err := uc.cardRepo.List(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("usecase list_cards: %w", err)
	}
	return dtos.ToCardResponses(cards), nil
}
