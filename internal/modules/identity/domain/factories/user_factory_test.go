package factories_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/factories"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/stretchr/testify/suite"
)

type UserFactorySuite struct{ suite.Suite }

func TestUserFactorySuite(t *testing.T) { suite.Run(t, new(UserFactorySuite)) }

func (s *UserFactorySuite) TestNew() {
	type args struct {
		name, email, pwd string
	}

	scenarios := []struct {
		name   string
		args   args
		expect func(err error)
	}{
		{
			name: "valid inputs",
			args: args{"Alice", "alice@example.com", "$2a$10$hash"},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "invalid email",
			args: args{"Alice", "not-email", "$2a$10$hash"},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidEmail)
			},
		},
		{
			name: "empty password",
			args: args{"Alice", "alice@example.com", ""},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidPassword)
			},
		},
		{
			name: "empty name",
			args: args{"", "alice@example.com", "$2a$10$hash"},
			expect: func(err error) {
				s.Error(err)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			u, err := factories.New(sc.args.name, sc.args.email, sc.args.pwd)
			sc.expect(err)
			if err == nil {
				s.NotNil(u)
				s.Equal(sc.args.name, u.Name())
			}
		})
	}
}
