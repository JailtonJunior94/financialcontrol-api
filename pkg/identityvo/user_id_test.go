package identityvo_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	"github.com/stretchr/testify/suite"
)

type UserIDSuite struct{ suite.Suite }

func TestUserIDSuite(t *testing.T) { suite.Run(t, new(UserIDSuite)) }

func (s *UserIDSuite) TestNewUserID() {
	id := identityvo.NewUserID()
	s.NotEmpty(id.String())
}

func (s *UserIDSuite) TestParseUserID() {
	scenarios := []struct {
		name   string
		input  string
		expect func(id identityvo.UserID, err error)
	}{
		{
			name:  "valid uuid-like string",
			input: "550e8400-e29b-41d4-a716-446655440000",
			expect: func(id identityvo.UserID, err error) {
				s.NoError(err)
				s.Equal(identityvo.UserID("550e8400-e29b-41d4-a716-446655440000"), id)
			},
		},
		{
			name:  "empty string",
			input: "",
			expect: func(id identityvo.UserID, err error) {
				s.Error(err)
				s.ErrorIs(err, identityvo.ErrInvalidUserID)
			},
		},
		{
			name:  "invalid non-uuid string",
			input: "not-a-uuid",
			expect: func(id identityvo.UserID, err error) {
				s.Error(err)
				s.ErrorIs(err, identityvo.ErrInvalidUserID)
			},
		},
		{
			name:  "valid uuid v1 is accepted (BUG-IDV-002 regression)",
			input: "c232ab00-9414-11ec-b3c8-9f6bdeced846",
			expect: func(id identityvo.UserID, err error) {
				s.NoError(err)
				s.Equal(identityvo.UserID("c232ab00-9414-11ec-b3c8-9f6bdeced846"), id)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			id, err := identityvo.ParseUserID(sc.input)
			sc.expect(id, err)
		})
	}
}

func (s *UserIDSuite) TestString() {
	id, _ := identityvo.ParseUserID("550e8400-e29b-41d4-a716-446655440000")
	s.Equal("550e8400-e29b-41d4-a716-446655440000", id.String())
}

func (s *UserIDSuite) TestNewUserIDIsUnique() {
	id1 := identityvo.NewUserID()
	id2 := identityvo.NewUserID()
	s.NotEqual(id1, id2)
}
