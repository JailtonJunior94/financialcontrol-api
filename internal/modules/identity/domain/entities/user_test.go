package entities_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"

	"github.com/stretchr/testify/suite"
)

type UserSuite struct{ suite.Suite }

func TestUserSuite(t *testing.T) { suite.Run(t, new(UserSuite)) }

func (s *UserSuite) TestNewUser() {
	email, _ := vos.NewEmail("user@example.com")
	pwd, _ := vos.NewHashedPassword("$2a$10$hash")

	type args struct {
		name  string
		email vos.Email
		pwd   vos.HashedPassword
	}

	scenarios := []struct {
		name   string
		args   args
		expect func(u *entities.User, err error)
	}{
		{
			name: "valid user",
			args: args{"John", email, pwd},
			expect: func(u *entities.User, err error) {
				s.NoError(err)
				s.NotNil(u)
				s.Equal("John", u.Name())
				s.Equal(email, u.Email())
				s.Equal(pwd, u.Password())
				s.True(u.Active())
				s.NotEmpty(u.ID().String())
			},
		},
		{
			name: "empty name",
			args: args{"", email, pwd},
			expect: func(u *entities.User, err error) {
				s.Error(err)
				s.Nil(u)
				s.ErrorIs(err, entities.ErrInvalidUser)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			u, err := entities.NewUser(sc.args.name, sc.args.email, sc.args.pwd)
			sc.expect(u, err)
		})
	}
}

func (s *UserSuite) TestRehydrate() {
	id, _ := vos.ParseUserID("550e8400-e29b-41d4-a716-446655440000")
	email, _ := vos.NewEmail("user@example.com")
	pwd, _ := vos.NewHashedPassword("$2a$10$hash")
	now := time.Now()

	u := entities.Rehydrate(id, "John", email, pwd, now, now, true)

	s.NotNil(u)
	s.Equal(id, u.ID())
	s.Equal("John", u.Name())
	s.Equal(email, u.Email())
	s.Equal(pwd, u.Password())
	s.Equal(now, u.CreatedAt())
	s.Equal(now, u.UpdatedAt())
	s.True(u.Active())
}
