package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func TestCreateCard_Execute(t *testing.T) {
	ctx := context.Background()
	userID := identityvo.NewUserID()

	tests := []struct {
		name   string
		req    dtos.CardRequest
		setup  func(*ifacemocks.CardRepository, *ifacemocks.FlagRepository)
		assert func(t *testing.T, out dtos.CardResponse, err error)
	}{
		{
			name: "sucesso",
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				reloaded := attachFlag(mustNewCard(userID), testFlagID, "Visa")
				cardRepo.EXPECT().Add(ctx, mock.MatchedBy(func(card *entities.Card) bool {
					return card.ExpirationDate().Equal(validCardRequest().ExpirationDate)
				})).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, mock.Anything).Return(reloaded, nil).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, out.ID)
				assert.Equal(t, testFlagID.String(), out.Flag.ID)
				assert.Equal(t, "Visa", out.Flag.Name)
			},
		},
		{
			name: "sucesso com closingDay igual ao dia de expirationDate preserva contrato legado",
			req: func() dtos.CardRequest {
				req := validCardRequest()
				req.ClosingDay = 10
				req.ExpirationDate = validCardRequest().ExpirationDate
				return req
			}(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				reloaded := attachFlag(mustLegacyCard(userID), testFlagID, "Visa")
				cardRepo.EXPECT().Add(ctx, mock.MatchedBy(func(card *entities.Card) bool {
					return card.ClosingDay().Int() == 10 && card.DueDay().Int() == 10
				})).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, mock.Anything).Return(reloaded, nil).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, 10, out.ClosingDay)
			},
		},
		{
			name:  "request inválido",
			req:   dtos.CardRequest{Name: "", ClosingDay: 0},
			setup: func(_ *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "flagId malformado é rejeitado antes de tocar o repositorio",
			req: func() dtos.CardRequest {
				r := validCardRequest()
				r.FlagID = "not-a-uuid"
				return r
			}(),
			setup: func(_ *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidFlagID)
				assert.Empty(t, out.ID)
			},
		},
		{
			name: "repo erro",
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().Add(ctx, mock.Anything).Return(errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "ErrFlagNotFound propagado do repo",
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().Add(ctx, mock.Anything).Return(domain.ErrFlagNotFound).Once()
			},
			assert: func(t *testing.T, out dtos.CardResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrFlagNotFound)
			},
		},
		{
			name: "erro ao recarregar cartao criado",
			req:  validCardRequest(),
			setup: func(cardRepo *ifacemocks.CardRepository, _ *ifacemocks.FlagRepository) {
				cardRepo.EXPECT().Add(ctx, mock.Anything).Return(nil).Once()
				cardRepo.EXPECT().GetByID(ctx, userID, mock.Anything).Return(nil, errors.New("reload error")).Once()
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
			sut := usecase.NewCreateCard(cardRepo, flagRepo)
			out, err := sut.Execute(ctx, userID, tt.req)
			tt.assert(t, out, err)
		})
	}
}
