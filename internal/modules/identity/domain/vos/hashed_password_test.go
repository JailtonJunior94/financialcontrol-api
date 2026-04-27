package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/stretchr/testify/suite"
)

type HashedPasswordSuite struct{ suite.Suite }

func TestHashedPasswordSuite(t *testing.T) { suite.Run(t, new(HashedPasswordSuite)) }

func (s *HashedPasswordSuite) TestNewHashedPassword() {
	scenarios := []struct {
		name   string
		input  string
		expect func(hp vos.HashedPassword, err error)
	}{
		{
			name:  "valid hash",
			input: "$2a$10$somehashvalue",
			expect: func(hp vos.HashedPassword, err error) {
				s.NoError(err)
				s.Equal(vos.HashedPassword("$2a$10$somehashvalue"), hp)
			},
		},
		{
			name:  "empty string",
			input: "",
			expect: func(hp vos.HashedPassword, err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidPassword)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			hp, err := vos.NewHashedPassword(sc.input)
			sc.expect(hp, err)
		})
	}
}

func (s *HashedPasswordSuite) TestString() {
	hp, _ := vos.NewHashedPassword("somehash")
	s.Equal("somehash", hp.String())
}
