package providers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cardsdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	cardsentities "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	cardsmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces/mocks"
	cardsvos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	financedomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	financevos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/providers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	testCardIDStr = "11111111-1111-4111-8111-111111111111"
	testUserIDStr = "22222222-2222-4222-8222-222222222222"
	testFlagIDStr = "33333333-3333-4333-8333-333333333333"
)

func makeTestCard(t *testing.T, active bool) *cardsentities.Card {
	t.Helper()
	flagID, err := cardsvos.ParseFlagID(testFlagIDStr)
	require.NoError(t, err)

	cardID, err := cardsvos.ParseCardID(testCardIDStr)
	require.NoError(t, err)

	userID, err := identityvo.ParseUserID(testUserIDStr)
	require.NoError(t, err)

	now := time.Now().UTC()
	expiration := time.Date(2026, time.December, 20, 0, 0, 0, 0, time.UTC)

	card, err := cardsentities.RehydrateCard(
		cardID,
		userID,
		flagID,
		"Nubank",
		"1234567890123456",
		"cartão principal",
		10,
		expiration,
		now,
		now,
		active,
	)
	require.NoError(t, err)
	return card
}

func TestCardProviderAdapter_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbErr := errors.New("db error")

	userID, err := identityvo.ParseUserID(testUserIDStr)
	require.NoError(t, err)

	cardsCardID, err := cardsvos.ParseCardID(testCardIDStr)
	require.NoError(t, err)

	financeCardID, err := financevos.ParseCardID(testCardIDStr)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupMock   func(t *testing.T, repo *cardsmocks.CardRepository)
		wantActive  bool
		wantFlag    string
		wantErr     error
		wantErrWrap bool
		wantEmpty   bool
	}{
		{
			name: "happy path — active card with flag",
			setupMock: func(t *testing.T, repo *cardsmocks.CardRepository) {
				card := makeTestCard(t, true)
				flagID, _ := cardsvos.ParseFlagID(testFlagIDStr)
				flag := cardsentities.NewFlag(flagID, "Visa", true)
				card.AttachFlag(&flag)
				repo.On("GetByID", ctx, userID, cardsCardID).Return(card, nil)
			},
			wantActive: true,
			wantFlag:   "Visa",
		},
		{
			name: "inactive card — Active=false, nil error (B2.a)",
			setupMock: func(t *testing.T, repo *cardsmocks.CardRepository) {
				card := makeTestCard(t, false)
				repo.On("GetByID", ctx, userID, cardsCardID).Return(card, nil)
			},
			wantActive: false,
			wantFlag:   "",
		},
		{
			name: "card not found or belongs to another user — ErrCardNotFound (B2.a)",
			setupMock: func(t *testing.T, repo *cardsmocks.CardRepository) {
				repo.On("GetByID", ctx, userID, cardsCardID).
					Return((*cardsentities.Card)(nil), cardsdomain.ErrCardNotFound)
			},
			wantErr:   financedomain.ErrCardNotFound,
			wantEmpty: true,
		},
		{
			name: "DB error — propagated wrapped",
			setupMock: func(t *testing.T, repo *cardsmocks.CardRepository) {
				repo.On("GetByID", ctx, userID, cardsCardID).
					Return((*cardsentities.Card)(nil), dbErr)
			},
			wantErrWrap: true,
			wantEmpty:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := cardsmocks.NewCardRepository(t)
			tc.setupMock(t, repo)

			adapter := providers.NewCardProviderAdapter(repo)
			got, err := adapter.GetByID(ctx, userID, financeCardID)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, projections.CardView{}, got)
				return
			}
			if tc.wantErrWrap {
				require.Error(t, err)
				assert.ErrorContains(t, err, "card provider")
				assert.Equal(t, projections.CardView{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, financeCardID, got.ID)
			assert.Equal(t, userID, got.UserID)
			assert.Equal(t, tc.wantFlag, got.FlagName)
			assert.Equal(t, 10, got.ClosingDay)
			assert.Equal(t, 20, got.DueDay)
			assert.Equal(t, 30, got.BillingCycle)
			assert.Equal(t, tc.wantActive, got.Active)
		})
	}
}
