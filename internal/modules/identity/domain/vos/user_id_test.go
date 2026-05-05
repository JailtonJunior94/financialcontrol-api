package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	sharedidentityvo "github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
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
			name:  "valid uuid v4",
			input: "550e8400-e29b-41d4-a716-446655440000",
			expect: func(id vos.UserID, err error) {
				s.NoError(err)
				s.Equal(vos.UserID("550e8400-e29b-41d4-a716-446655440000"), id)
			},
		},
		{
			name:  "valid uuid v1 is accepted (BUG-IDV-002 regression)",
			input: "c232ab00-9414-11ec-b3c8-9f6bdeced846",
			expect: func(id vos.UserID, err error) {
				s.NoError(err)
				s.Equal(vos.UserID("c232ab00-9414-11ec-b3c8-9f6bdeced846"), id)
			},
		},
		{
			name:  "empty string",
			input: "",
			expect: func(id vos.UserID, err error) {
				s.ErrorIs(err, vos.ErrInvalidUserID)
			},
		},
		{
			name:  "invalid non-uuid string",
			input: "not-a-uuid",
			expect: func(id vos.UserID, err error) {
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

func (s *UserIDSuite) TestSharedIdentityVOCompatibility() {
	sharedID, err := sharedidentityvo.ParseUserID("550e8400-e29b-41d4-a716-446655440000")
	s.Require().NoError(err)

	legacyID := sharedID

	s.Equal(sharedID.String(), legacyID.String())
	s.Equal(sharedidentityvo.ErrInvalidUserID, vos.ErrInvalidUserID)
}
