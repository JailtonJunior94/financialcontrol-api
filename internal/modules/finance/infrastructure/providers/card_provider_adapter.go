package providers

import (
	"context"
	"errors"
	"fmt"

	cardsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	cardsinterfaces "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports"
	cardsvos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	financedomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	financevos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var _ ports.CardProvider = (*CardProviderAdapter)(nil)

// CardProviderAdapter wraps cards.CardRepository to implement finance.ports.CardProvider.
// Decisão B2.a: not-found or wrong-user → ErrCardNotFound; inactive card → CardView{Active:false}, nil.
type CardProviderAdapter struct {
	repo cardsinterfaces.CardRepository
}

func NewCardProviderAdapter(repo cardsinterfaces.CardRepository) *CardProviderAdapter {
	return &CardProviderAdapter{repo: repo}
}

func (a *CardProviderAdapter) GetByID(
	ctx context.Context,
	userID identityvo.UserID,
	cardID financevos.CardID,
) (projections.CardView, error) {
	cid, err := cardsvos.ParseCardID(cardID.String())
	if err != nil {
		return projections.CardView{}, fmt.Errorf("card provider: parse card id: %w", err)
	}

	card, err := a.repo.GetByID(ctx, userID, cid)
	if errors.Is(err, cardsdomain.ErrCardNotFound) {
		return projections.CardView{}, financedomain.ErrCardNotFound
	}
	if err != nil {
		return projections.CardView{}, fmt.Errorf("card provider: get by id: %w", err)
	}

	flagName := ""
	if f := card.Flag(); f != nil {
		flagName = f.Name()
	}

	return projections.CardView{
		ID:           financevos.CardID(card.ID().String()),
		UserID:       card.UserID(),
		FlagName:     flagName,
		ClosingDay:   card.ClosingDay().Int(),
		DueDay:       card.DueDay().Int(),
		BillingCycle: 30,
		Active:       card.Active(),
	}, nil
}
