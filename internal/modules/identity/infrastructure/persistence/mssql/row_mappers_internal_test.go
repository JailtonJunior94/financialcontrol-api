package mssql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowToEntity(t *testing.T) {
	t.Parallel()

	validID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	createdAt := time.Date(2024, time.March, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, time.April, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		row     userRow
		wantErr bool
	}{
		{
			name:    "valid row maps to user",
			row:     userRow{ID: validID, Name: "Jailton", Email: "user@example.com", Password: "$2a$10$hash", CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: false,
		},
		{
			name:    "invalid id returns error",
			row:     userRow{ID: "not-a-uuid", Name: "Jailton", Email: "user@example.com", Password: "$2a$10$hash", CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: true,
		},
		{
			name:    "invalid email returns error",
			row:     userRow{ID: validID, Name: "Jailton", Email: "not-an-email", Password: "$2a$10$hash", CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: true,
		},
		{
			name:    "empty password returns error",
			row:     userRow{ID: validID, Name: "Jailton", Email: "user@example.com", Password: "", CreatedAt: createdAt, UpdatedAt: updatedAt, Active: true},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			user, err := rowToEntity(&tc.row)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.row.Email, user.Email().String())
			assert.Equal(t, tc.row.Active, user.Active())
		})
	}
}
