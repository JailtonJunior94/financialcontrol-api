package mssql_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/persistence/mssql"
)

func TestRowToCategory(t *testing.T) {
	t.Parallel()

	validID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	createdAt := time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, time.April, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		row     mssql.CategoryRow
		wantErr bool
	}{
		{
			name:    "valid row maps to category",
			row:     mssql.CategoryRow{ID: validID, Name: "Alimentação", Sequence: 1, CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: false,
		},
		{
			name:    "invalid id returns error",
			row:     mssql.CategoryRow{ID: "not-a-uuid", Name: "Alimentação", Sequence: 1, CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: true,
		},
		{
			name:    "empty name returns error",
			row:     mssql.CategoryRow{ID: validID, Name: "", Sequence: 1, CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cat, err := mssql.RowToCategory(&tc.row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.row.Name, cat.Name().String())
			assert.Equal(t, tc.row.Active, cat.IsActive())
		})
	}
}
