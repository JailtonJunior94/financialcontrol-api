package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func TestDeactivateCard_Execute(t *testing.T) {
	ctx := context.Background()
	userID := identityvo.NewUserID()
	card := mustNewCard(userID)
	cardID := card.ID()

	tests := []struct {
		name   string
		userID identityvo.UserID
		id     vos.CardID
		setup  func(*ifacemocks.CardRepository)
		assert func(t *testing.T, err error)
	}{
		{
			name:   "sucesso",
			userID: userID,
			id:     cardID,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(card, nil).Once()
				cardRepo.EXPECT().Update(ctx, card).Return(nil).Once()
			},
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "card inexistente (repo retorna ErrCardNotFound)",
			userID: userID,
			id:     cardID,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(nil, domain.ErrCardNotFound).Once()
			},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrCardNotFound)
			},
		},
		{
			name:   "idempotencia: card ja inativo nao chama Update",
			userID: userID,
			id:     cardID,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				inactiveCard := mustNewCard(userID)
				inactiveCard.Deactivate()
				require.False(t, inactiveCard.Active())
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(inactiveCard, nil).Once()
				// Update NAO deve ser chamado: a expectativa ausente garante isso pelo mockery.
			},
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "repo GetByID erro",
			userID: userID,
			id:     cardID,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := ifacemocks.NewCardRepository(t)
			tt.setup(cardRepo)
			sut := usecase.NewDeactivateCard(cardRepo)
			err := sut.Execute(ctx, tt.userID, tt.id)
			tt.assert(t, err)
		})
	}
}
