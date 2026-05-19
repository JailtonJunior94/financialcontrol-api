package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func TestUpdateCard_Execute(t *testing.T) {
	ctx := context.Background()
	userID := identityvo.NewUserID()
	card := attachFlag(mustNewCard(userID), testFlagID, "Visa")
	cardID := card.ID()
	legacyCard := attachFlag(mustLegacyCard(userID), testFlagID, "Visa")
	legacyCardID := legacyCard.ID()
	updatedReq := validCardRequestWithFlag(alternateFlagID)

	tests := []struct {
		name   string
		id     vos.CardID
		req    dtos.CardRequest
		setup  func(*ifacemocks.CardRepository, *ifacemocks.FlagRepository)
		assert func(t *testing.T, out dtos.CardResponse, err error)
	}{
		{
			name: "sucesso",
			id:   cardID,
			req:  updatedReq,
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				reloaded := attachFlag(mustRehydrateCardWithID(userID, cardID), alternateFlagID, "Mastercard")
				require.NoError(t, reloaded.Update(alternateFlagID, updatedReq.Name, updatedReq.Number, updatedReq.Description, updatedReq.ClosingDay, updatedReq.ExpirationDate))
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(card, nil).Once()
				cardRepo.EXPECT().Update(ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ExpirationDate().Equal(updatedReq.ExpirationDate)
				})).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(reloaded, nil).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, cardID.String(), out.ID)
				assert.Equal(t, alternateFlagID.String(), out.Flag.ID)
				assert.Equal(t, "Mastercard", out.Flag.Name)
			},
		},
		{
			name: "cartao legado com closing igual due continua atualizavel quando ciclo nao muda",
			id:   legacyCardID,
			req: dtos.CardRequest{
				FlagID:         testFlagID.String(),
				Name:           "Cartão Legado Atualizado",
				Number:         "4111111111111111",
				Description:    "desc atualizado",
				ClosingDay:     10,
				ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			},
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				reloaded := attachFlag(mustLegacyCardWithID(userID, legacyCardID), testFlagID, "Visa")
				require.NoError(t, reloaded.Update(testFlagID, "Cartão Legado Atualizado", "4111111111111111", "desc atualizado", 10, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)))
				cardRepo.EXPECT().GetByID(ctx, userID, legacyCardID).Return(legacyCard, nil).Once()
				cardRepo.EXPECT().Update(ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ClosingDay().Int() == 10 &&
						updated.DueDay().Int() == 10 &&
						updated.Name().String() == "Cartão Legado Atualizado"
				})).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, legacyCardID).Return(reloaded, nil).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Cartão Legado Atualizado", out.Name)
				assert.Equal(t, 10, out.ClosingDay)
			},
		},
		{
			name: "cartao moderno continua atualizavel para closing igual due para preservar contrato publico",
			id:   cardID,
			req: dtos.CardRequest{
				FlagID:         testFlagID.String(),
				Name:           "Cartão Mesmo Dia",
				Number:         "4111111111111111",
				Description:    "desc mesmo dia",
				ClosingDay:     10,
				ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			},
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				reloaded := attachFlag(mustRehydrateCardWithID(userID, cardID), testFlagID, "Visa")
				require.NoError(t, reloaded.Update(testFlagID, "Cartão Mesmo Dia", "4111111111111111", "desc mesmo dia", 10, time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)))
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(card, nil).Once()
				cardRepo.EXPECT().Update(ctx, mock.MatchedBy(func(updated *entities.Card) bool {
					return updated.ClosingDay().Int() == 10 &&
						updated.DueDay().Int() == 10 &&
						updated.Name().String() == "Cartão Mesmo Dia"
				})).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(reloaded, nil).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Cartão Mesmo Dia", out.Name)
				assert.Equal(t, 10, out.ClosingDay)
			},
		},
		{
			name:  "request inválido",
			id:    cardID,
			req:   dtos.CardRequest{},
			setup: func(_ *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "card inexistente (repo retorna ErrCardNotFound)",
			id:   cardID,
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(nil, domain.ErrCardNotFound).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrCardNotFound)
			},
		},
		{
			name: "flag inexistente via repo",
			id:   cardID,
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(card, nil).Once()
				cardRepo.EXPECT().Update(ctx, mock.Anything).Return(domain.ErrFlagNotFound).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrFlagNotFound)
			},
		},
		{
			name: "repo GetByID erro",
			id:   cardID,
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "erro ao recarregar cartao atualizado",
			id:   cardID,
			req:  updatedReq,
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(card, nil).Once()
				cardRepo.EXPECT().Update(ctx, mock.Anything).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, cardID).Return(nil, errors.New("reload error")).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := ifacemocks.NewCardRepository(t)
			flagRepo := ifacemocks.NewFlagRepository(t)
			tt.setup(cardRepo, flagRepo)
			sut := usecase.NewUpdateCard(cardRepo, flagRepo)
			out, err := sut.Execute(ctx, userID, tt.id, tt.req)
			tt.assert(t, out, err)
		})
	}
}
