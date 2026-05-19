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
)

func TestListFlags_Execute(t *testing.T) {
	ctx := context.Background()
	flag := entities.NewFlag(testFlagID, "Visa", true)

	tests := []struct {
		name   string
		setup  func(*ifacemocks.FlagRepository)
		assert func(t *testing.T, out []dtos.FlagResponse, err error)
	}{
		{
			name: "sucesso sempre retorna lista completa",
			setup: func(flagRepo *ifacemocks.FlagRepository) {
				flagRepo.EXPECT().List(ctx).Return([]entities.Flag{flag}, nil).Once()
			},
			assert: func(t *testing.T, out []dtos.FlagResponse, err error) {
				require.NoError(t, err)
				assert.Len(t, out, 1)
				assert.Equal(t, "Visa", out[0].Name)
			},
		},
		{
			name: "repo erro",
			setup: func(flagRepo *ifacemocks.FlagRepository) {
				flagRepo.EXPECT().List(ctx).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out []dtos.FlagResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagRepo := ifacemocks.NewFlagRepository(t)
			tt.setup(flagRepo)
			sut := usecase.NewListFlags(flagRepo)
			out, err := sut.Execute(ctx)
			tt.assert(t, out, err)
		})
	}
}
