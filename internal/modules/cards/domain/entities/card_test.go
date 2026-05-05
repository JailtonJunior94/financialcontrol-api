package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	validUserIDStr = "a0000000-0000-4000-8000-000000000001"
	validFlagIDStr = "b0000000-0000-4000-8000-000000000002"
	validName      = "Nubank"
	validNumber    = "1234567890123456"
	validDesc      = "cartão principal"
	validClosing   = 10
	validDueDay    = 20
)

func validExpiration() time.Time {
	return time.Date(2026, time.May, validDueDay, 0, 0, 0, 0, time.UTC)
}

func expirationOnDay(day int) time.Time {
	return time.Date(2026, time.May, day, 0, 0, 0, 0, time.UTC)
}

func mustUserID(t *testing.T) identityvo.UserID {
	t.Helper()
	id, err := identityvo.ParseUserID(validUserIDStr)
	require.NoError(t, err)
	return id
}

func mustFlagID(t *testing.T) vos.FlagID {
	t.Helper()
	id, err := vos.ParseFlagID(validFlagIDStr)
	require.NoError(t, err)
	return id
}

func TestNewCard_HappyPath(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, validClosing, validExpiration())
	require.NoError(t, err)

	assert.NotEmpty(t, card.ID())
	assert.Equal(t, mustUserID(t), card.UserID())
	assert.Equal(t, mustFlagID(t), card.FlagID())
	assert.Equal(t, validName, card.Name().String())
	assert.Equal(t, validNumber, card.Number().String())
	assert.Equal(t, validDesc, card.Description())
	assert.Equal(t, validClosing, card.ClosingDay().Int())
	assert.Equal(t, validDueDay, card.DueDay().Int())
	assert.True(t, card.ExpirationDate().Equal(validExpiration()))
	assert.True(t, card.Active())
	assert.False(t, card.CreatedAt().IsZero())
}

func TestNewCard_Invariants(t *testing.T) {
	t.Parallel()

	userID := mustUserID(t)
	flagID := mustFlagID(t)

	cases := []struct {
		name        string
		cardName    string
		number      string
		closingDay  int
		expiration  time.Time
		expectedErr error
	}{
		{"invalid name empty", "", validNumber, validClosing, validExpiration(), vos.ErrInvalidCardName},
		{"invalid name too long", string(make([]byte, 101)), validNumber, validClosing, validExpiration(), vos.ErrInvalidCardName},
		{"invalid number too short", validName, "12345678901", validClosing, validExpiration(), vos.ErrInvalidCardNumber},
		{"invalid number too long", validName, "12345678901234567890", validClosing, validExpiration(), vos.ErrInvalidCardNumber},
		{"invalid number with letters", validName, "12345678901abc", validClosing, validExpiration(), vos.ErrInvalidCardNumber},
		{"invalid closing day 0", validName, validNumber, 0, validExpiration(), vos.ErrInvalidClosingDay},
		{"invalid closing day 32", validName, validNumber, 32, validExpiration(), vos.ErrInvalidClosingDay},
		{"expiration date zero", validName, validNumber, validClosing, time.Time{}, vos.ErrInvalidDueDay},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := entities.NewCard(userID, flagID, tc.cardName, tc.number, validDesc, tc.closingDay, tc.expiration)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)
		})
	}
}

func TestNewCard_SameDayCyclePreservesLegacyContract(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, 10, expirationOnDay(10))
	require.NoError(t, err)

	assert.Equal(t, 10, card.ClosingDay().Int())
	assert.Equal(t, 10, card.DueDay().Int())
	assert.True(t, card.ExpirationDate().Equal(expirationOnDay(10)))
}

func TestCard_Update(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, validClosing, validExpiration())
	require.NoError(t, err)

	originalCreatedAt := card.CreatedAt()
	originalID := card.ID()
	originalUserID := card.UserID()

	newFlagID, _ := vos.ParseFlagID("c0000000-0000-4000-8000-000000000003")
	err = card.Update(newFlagID, "Itaú", "9876543210123456", "updated", 5, expirationOnDay(15))
	require.NoError(t, err)

	assert.Equal(t, originalID, card.ID())
	assert.Equal(t, originalUserID, card.UserID())
	assert.Equal(t, originalCreatedAt, card.CreatedAt())
	assert.Equal(t, newFlagID, card.FlagID())
	assert.Equal(t, "Itaú", card.Name().String())
	assert.Equal(t, 5, card.ClosingDay().Int())
	assert.Equal(t, 15, card.DueDay().Int())
	assert.True(t, card.ExpirationDate().Equal(expirationOnDay(15)))
	assert.True(t, card.Active())
}

func TestCard_Update_LegacySameDayCyclePreservesCompatibility(t *testing.T) {
	t.Parallel()

	cardID := vos.NewCardID()
	expirationDate := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	card, err := entities.RehydrateCard(
		cardID,
		mustUserID(t),
		mustFlagID(t),
		validName,
		validNumber,
		validDesc,
		10,
		expirationDate,
		time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		true,
	)
	require.NoError(t, err)

	err = card.Update(mustFlagID(t), "Legacy Updated", validNumber, "updated", 10, expirationOnDay(10))
	require.NoError(t, err)
	assert.Equal(t, "Legacy Updated", card.Name().String())
	assert.Equal(t, 10, card.ClosingDay().Int())
	assert.Equal(t, 10, card.DueDay().Int())
}

func TestCard_Update_InvalidFields(t *testing.T) {
	t.Parallel()

	flagID := mustFlagID(t)

	cases := []struct {
		name        string
		cardName    string
		number      string
		closingDay  int
		expiration  time.Time
		expectedErr error
	}{
		{"invalid name", "", validNumber, validClosing, validExpiration(), vos.ErrInvalidCardName},
		{"invalid closing day", validName, validNumber, 0, validExpiration(), vos.ErrInvalidClosingDay},
		{"expiration zero", validName, validNumber, validClosing, time.Time{}, vos.ErrInvalidDueDay},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card, err := entities.NewCard(mustUserID(t), flagID, validName, validNumber, validDesc, validClosing, validExpiration())
			require.NoError(t, err)
			err = card.Update(flagID, tc.cardName, tc.number, validDesc, tc.closingDay, tc.expiration)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tc.expectedErr))
		})
	}
}

func TestCard_Update_SameDayCyclePreservesLegacyContract(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, validClosing, validExpiration())
	require.NoError(t, err)

	err = card.Update(mustFlagID(t), "Mesmo Dia", validNumber, validDesc, 10, expirationOnDay(10))
	require.NoError(t, err)

	assert.Equal(t, 10, card.ClosingDay().Int())
	assert.Equal(t, 10, card.DueDay().Int())
	assert.True(t, card.ExpirationDate().Equal(expirationOnDay(10)))
}

func TestCard_Deactivate(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, validClosing, validExpiration())
	require.NoError(t, err)
	assert.True(t, card.Active())

	card.Deactivate()
	assert.False(t, card.Active())
	assert.False(t, card.UpdatedAt().IsZero())
}

func TestCard_BestPurchaseDay(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, 10, expirationOnDay(20))
	require.NoError(t, err)
	assert.Equal(t, 9, card.BestPurchaseDay())
}

func TestCard_DueDateFor(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, 25, expirationOnDay(1))
	require.NoError(t, err)

	purchase := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	due := card.DueDateFor(purchase)
	assert.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), due)
}

func TestCard_DueDateFor_OnClosingDayMovesToNextCycle(t *testing.T) {
	t.Parallel()

	card, err := entities.NewCard(mustUserID(t), mustFlagID(t), validName, validNumber, validDesc, 25, expirationOnDay(1))
	require.NoError(t, err)

	purchase := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	due := card.DueDateFor(purchase)
	assert.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), due)
}
