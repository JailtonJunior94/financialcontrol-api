package dtos_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

var validFlagID = "550e8400-e29b-41d4-a716-446655440000"

func validRequest() dtos.CardRequest {
	return dtos.CardRequest{
		FlagID:         validFlagID,
		Name:           "Meu Cartão",
		Number:         "4111111111111111",
		Description:    "desc",
		ClosingDay:     15,
		ExpirationDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
	}
}

func TestToCardResponse_PreservesExpirationDate(t *testing.T) {
	t.Parallel()

	userID, err := identityvo.ParseUserID("a0000000-0000-4000-8000-000000000001")
	require.NoError(t, err)

	flagID, err := vos.ParseFlagID(validFlagID)
	require.NoError(t, err)

	expirationDate := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	card, err := entities.NewCard(userID, flagID, "Meu Cartão", "4111111111111111", "desc", 15, expirationDate)
	require.NoError(t, err)

	resp := dtos.ToCardResponse(card)
	assert.True(t, resp.ExpirationDate.Equal(expirationDate))
}

func TestCardRequest_Validate(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name      string
		req       dtos.CardRequest
		expectErr bool
		errIs     error
	}{
		{
			name:      "válido",
			req:       validRequest(),
			expectErr: false,
		},
		{
			name: "flagId inválido",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.FlagID = "not-a-uuid"
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidFlagID,
		},
		{
			name: "nome vazio",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.Name = ""
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidCardName,
		},
		{
			name: "número inválido",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.Number = "123"
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidCardNumber,
		},
		{
			name: "closingDay zero",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.ClosingDay = 0
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidClosingDay,
		},
		{
			name: "closingDay maior que 31",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.ClosingDay = 32
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidClosingDay,
		},
		{
			name: "expirationDate zero",
			req: func() dtos.CardRequest {
				r := validRequest()
				r.ExpirationDate = time.Time{}
				return r
			}(),
			expectErr: true,
			errIs:     domain.ErrInvalidDueDay,
		},
	}

	for _, tc := range scenarios {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.req.Validate()
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errIs != nil {
					assert.ErrorIs(t, err, tc.errIs)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}
