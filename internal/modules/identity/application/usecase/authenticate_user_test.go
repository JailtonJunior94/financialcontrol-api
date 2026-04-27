package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type AuthenticateUserSuite struct {
	suite.Suite
	ctx         context.Context
	userRepo    *ifacemocks.UserRepository
	hasher      *ifacemocks.Hasher
	tokenIssuer *ifacemocks.TokenIssuer
	sut         usecase.AuthenticateUser
}

func TestAuthenticateUserSuite(t *testing.T) { suite.Run(t, new(AuthenticateUserSuite)) }

func (s *AuthenticateUserSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepo = ifacemocks.NewUserRepository(s.T())
	s.hasher = ifacemocks.NewHasher(s.T())
	s.tokenIssuer = ifacemocks.NewTokenIssuer(s.T())
	s.sut = usecase.NewAuthenticateUser(s.userRepo, s.hasher, s.tokenIssuer)
}

func (s *AuthenticateUserSuite) TestExecute() {
	email, _ := vos.NewEmail("user@example.com")
	pwd, _ := vos.NewHashedPassword("hashedpwd")
	user := entities.Rehydrate(vos.NewUserID(), "Test User", email, pwd, time.Now(), time.Now(), true)
	expiresAt := time.Now().Add(15 * time.Minute)

	type args struct{ in dtos.AuthRequest }
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(out dtos.AuthResponse, err error)
	}{
		{
			name: "sucesso",
			args: args{in: dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(user, nil).Once()
				s.hasher.EXPECT().Verify(pwd, "plainpwd").Return(true).Once()
				s.tokenIssuer.EXPECT().Issue(s.ctx, user.ID(), user.Email()).Return("token123", expiresAt, nil).Once()
			},
			expect: func(out dtos.AuthResponse, err error) {
				s.NoError(err)
				s.Equal("token123", out.Token)
				s.Equal(expiresAt, out.ExpiresAt)
			},
		},
		{
			name:  "e-mail invalido",
			args:  args{in: dtos.AuthRequest{Email: "not-an-email", Password: "plainpwd"}},
			setup: func() {},
			expect: func(out dtos.AuthResponse, err error) {
				s.ErrorIs(err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "e-mail desconhecido",
			args: args{in: dtos.AuthRequest{Email: "unknown@example.com", Password: "plainpwd"}},
			setup: func() {
				unknownEmail, _ := vos.NewEmail("unknown@example.com")
				s.userRepo.EXPECT().GetByEmail(s.ctx, unknownEmail).Return(nil, nil).Once()
			},
			expect: func(out dtos.AuthResponse, err error) {
				s.ErrorIs(err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "senha incorreta",
			args: args{in: dtos.AuthRequest{Email: "user@example.com", Password: "wrong"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(user, nil).Once()
				s.hasher.EXPECT().Verify(pwd, "wrong").Return(false).Once()
			},
			expect: func(out dtos.AuthResponse, err error) {
				s.ErrorIs(err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "falha de repo",
			args: args{in: dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out dtos.AuthResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "falha de issuer",
			args: args{in: dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(user, nil).Once()
				s.hasher.EXPECT().Verify(pwd, "plainpwd").Return(true).Once()
				s.tokenIssuer.EXPECT().Issue(s.ctx, user.ID(), user.Email()).Return("", time.Time{}, errors.New("issuer error")).Once()
			},
			expect: func(out dtos.AuthResponse, err error) {
				s.ErrorIs(err, domain.ErrTokenIssuance)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(s.ctx, sc.args.in)
			sc.expect(out, err)
		})
	}
}
