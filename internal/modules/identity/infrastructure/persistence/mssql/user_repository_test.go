package mssql

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/stretchr/testify/suite"
)

type RowConversionSuite struct{ suite.Suite }

func TestRowConversionSuite(t *testing.T) { suite.Run(t, new(RowConversionSuite)) }

func (s *RowConversionSuite) TestRowToEntity() {
	now := time.Now()

	scenarios := []struct {
		name   string
		row    *userRow
		expect func(u *entities.User, err error)
	}{
		{
			name: "valid row maps to entity",
			row: &userRow{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Alice",
				Email:     "alice@example.com",
				Password:  "$2a$10$somehash",
				CreatedAt: now,
				UpdatedAt: now,
				Active:    true,
			},
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(u)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", u.ID().String())
				s.Equal("Alice", u.Name())
				s.Equal("alice@example.com", u.Email().String())
				s.Equal("$2a$10$somehash", u.Password().String())
				s.Equal(now, u.CreatedAt())
				s.Equal(now, u.UpdatedAt())
				s.True(u.Active())
			},
		},
		{
			name: "inactive user",
			row: &userRow{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Bob",
				Email:     "bob@example.com",
				Password:  "$2a$10$somehash",
				CreatedAt: now,
				UpdatedAt: now,
				Active:    false,
			},
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(u)
				s.False(u.Active())
			},
		},
		{
			name: "empty user ID returns error",
			row: &userRow{
				ID:        "",
				Name:      "Charlie",
				Email:     "charlie@example.com",
				Password:  "$2a$10$somehash",
				CreatedAt: now,
				UpdatedAt: now,
				Active:    true,
			},
			expect: func(u *entities.User, err error) {
				s.Require().Error(err)
				s.Nil(u)
			},
		},
		{
			name: "invalid email returns error",
			row: &userRow{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Dave",
				Email:     "not-an-email",
				Password:  "$2a$10$somehash",
				CreatedAt: now,
				UpdatedAt: now,
				Active:    true,
			},
			expect: func(u *entities.User, err error) {
				s.Require().Error(err)
				s.Nil(u)
			},
		},
		{
			name: "empty password returns error",
			row: &userRow{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Eve",
				Email:     "eve@example.com",
				Password:  "",
				CreatedAt: now,
				UpdatedAt: now,
				Active:    true,
			},
			expect: func(u *entities.User, err error) {
				s.Require().Error(err)
				s.Nil(u)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			u, err := rowToEntity(sc.row)
			sc.expect(u, err)
		})
	}
}
