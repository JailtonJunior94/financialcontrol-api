package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func TestListCards_Execute(t *testing.T) {
	ctx := context.Background()
	userID := identityvo.NewUserID()
	card := mustNewCard(userID)

	unpaged := dtos.Pagination{}
	clampedPagination := dtos.NewPagination(1, 1000)

	tests := []struct {
		name       string
		pagination dtos.Pagination
		setup      func(*ifacemocks.CardRepository)
		assert     func(t *testing.T, out []dtos.CardResponse, err error)
	}{
		{
			name:       "sucesso sem paginacao preserva contrato legado",
			pagination: unpaged,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().List(ctx, userID, unpaged).Return([]entities.Card{*card}, nil).Once()
			},
			assert: func(t *testing.T, out []dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Len(t, out, 1)
			},
		},
		{
			name:       "paginacao zero nao e normalizada para truncar a lista",
			pagination: dtos.Pagination{Page: 0, Size: 0},
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().List(ctx, userID, unpaged).Return([]entities.Card{*card}, nil).Once()
			},
			assert: func(t *testing.T, out []dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Len(t, out, 1)
			},
		},
		{
			name:       "size acima do max e clampado",
			pagination: dtos.Pagination{Page: 1, Size: 1000},
			setup: func(cardRepo *ifacemocks.CardRepository) {
				assert.Equal(t, 200, clampedPagination.Size)
				cardRepo.EXPECT().List(ctx, userID, clampedPagination).Return([]entities.Card{*card}, nil).Once()
			},
			assert: func(t *testing.T, out []dtos.CardResponse, err error) {
				require.NoError(t, err)
				assert.Len(t, out, 1)
			},
		},
		{
			name:       "repo retorna erro",
			pagination: unpaged,
			setup: func(cardRepo *ifacemocks.CardRepository) {
				cardRepo.EXPECT().List(ctx, userID, unpaged).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out []dtos.CardResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := ifacemocks.NewCardRepository(t)
			tt.setup(cardRepo)
			sut := usecase.NewListCards(cardRepo)
			out, err := sut.Execute(ctx, userID, tt.pagination)
			tt.assert(t, out, err)
		})
	}
}
