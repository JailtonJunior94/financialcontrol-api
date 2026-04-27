package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/stretchr/testify/suite"
)

type UserIDSuite struct{ suite.Suite }

func TestUserIDSuite(t *testing.T) { suite.Run(t, new(UserIDSuite)) }

func (s *UserIDSuite) TestNewUserID() {
	id := vos.NewUserID()
	s.NotEmpty(id.String())
}

func (s *UserIDSuite) TestParseUserID() {
	scenarios := []struct {
		name   string
		input  string
		expect func(id vos.UserID, err error)
	}{
		{
			name:  "valid uuid-like string",
			input: "550e8400-e29b-41d4-a716-446655440000",
			expect: func(id vos.UserID, err error) {
				s.NoError(err)
				s.Equal(vos.UserID("550e8400-e29b-41d4-a716-446655440000"), id)
			},
		},
		{
			name:  "empty string",
			input: "",
			expect: func(id vos.UserID, err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidUserID)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			id, err := vos.ParseUserID(sc.input)
			sc.expect(id, err)
		})
	}
}

func (s *UserIDSuite) TestString() {
	id, _ := vos.ParseUserID("550e8400-e29b-41d4-a716-446655440000")
	s.Equal("550e8400-e29b-41d4-a716-446655440000", id.String())
}

func (s *UserIDSuite) TestNewUserIDIsUnique() {
	id1 := vos.NewUserID()
	id2 := vos.NewUserID()
	s.NotEqual(id1, id2)
}
