package mssql_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
)

func TestRowToCard(t *testing.T) {
	t.Parallel()

	validCardID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	validUserID := "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e"
	validFlagID := "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f"
	validFlagEID := "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f"
	expDate := time.Date(1900, time.January, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, time.April, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		row     mssql.CardRow
		wantErr bool
		check   func(t *testing.T, row mssql.CardRow, card interface{})
	}{
		{
			name: "valid row maps to card",
			row: mssql.CardRow{
				ID:             validCardID,
				UserID:         validUserID,
				FlagID:         validFlagID,
				Name:           "Meu Cartão",
				Number:         "123456789012",
				Description:    "Descrição",
				ClosingDay:     10,
				ExpirationDate: expDate,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				Active:         true,
				FlagEntityID:   validFlagEID,
				FlagName:       "Visa",
				FlagActive:     true,
			},
			wantErr: false,
		},
		{
			name: "invalid userID returns error",
			row: mssql.CardRow{
				ID:             validCardID,
				UserID:         "not-a-uuid",
				FlagID:         validFlagID,
				Name:           "Meu Cartão",
				Number:         "123456789012",
				Description:    "",
				ClosingDay:     10,
				ExpirationDate: expDate,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				Active:         true,
				FlagEntityID:   validFlagEID,
				FlagName:       "Visa",
				FlagActive:     true,
			},
			wantErr: true,
		},
		{
			name: "invalid flagID returns error",
			row: mssql.CardRow{
				ID:             validCardID,
				UserID:         validUserID,
				FlagID:         "not-a-uuid",
				Name:           "Meu Cartão",
				Number:         "123456789012",
				Description:    "",
				ClosingDay:     10,
				ExpirationDate: expDate,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				Active:         true,
				FlagEntityID:   validFlagEID,
				FlagName:       "Visa",
				FlagActive:     true,
			},
			wantErr: true,
		},
		{
			name: "invalid cardID returns error",
			row: mssql.CardRow{
				ID:             "bad-id",
				UserID:         validUserID,
				FlagID:         validFlagID,
				Name:           "Meu Cartão",
				Number:         "123456789012",
				Description:    "",
				ClosingDay:     10,
				ExpirationDate: expDate,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				Active:         true,
				FlagEntityID:   validFlagEID,
				FlagName:       "Visa",
				FlagActive:     true,
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card, err := mssql.RowToCard(&tc.row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.row.Name, card.Name().String())
			assert.Equal(t, tc.row.Number, card.Number().String())
			assert.Equal(t, tc.row.Description, card.Description())
			assert.Equal(t, tc.row.ClosingDay, card.ClosingDay().Int())
			assert.Equal(t, tc.row.ExpirationDate.Day(), card.DueDay().Int())
			assert.True(t, tc.row.ExpirationDate.Equal(card.ExpirationDate()))
			assert.NotNil(t, card.Flag())
			assert.Equal(t, tc.row.FlagName, card.Flag().Name())
		})
	}
}

func TestRowToCard_AllowsLegacyClosingEqualsDue(t *testing.T) {
	t.Parallel()

	expirationDate := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	card, err := mssql.RowToCard(&mssql.CardRow{
		ID:             "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d",
		UserID:         "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
		FlagID:         "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
		Name:           "Meu Cartão",
		Number:         "123456789012",
		Description:    "Descrição",
		ClosingDay:     10,
		ExpirationDate: expirationDate,
		CreatedAt:      time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2024, time.April, 1, 12, 0, 0, 0, time.UTC),
		Active:         true,
		FlagEntityID:   "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
		FlagName:       "Visa",
		FlagActive:     true,
	})
	require.NoError(t, err)
	assert.Equal(t, 10, card.ClosingDay().Int())
	assert.Equal(t, 10, card.DueDay().Int())
	assert.True(t, card.ExpirationDate().Equal(expirationDate))
}

func TestRowToFlag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		row     mssql.FlagRow
		wantErr bool
	}{
		{
			name:    "valid row maps to flag",
			row:     mssql.FlagRow{ID: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f", Name: "Visa", Active: true},
			wantErr: false,
		},
		{
			name:    "invalid id returns error",
			row:     mssql.FlagRow{ID: "not-a-uuid", Name: "Visa", Active: true},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			flag, err := mssql.RowToFlag(&tc.row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.row.Name, flag.Name())
			assert.Equal(t, tc.row.Active, flag.Active())
		})
	}
}
