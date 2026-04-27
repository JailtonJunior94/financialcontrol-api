package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type CreateUserSuite struct {
	suite.Suite
	ctx      context.Context
	userRepo *ifacemocks.UserRepository
	hasher   *ifacemocks.Hasher
	sut      usecase.CreateUser
}

func TestCreateUserSuite(t *testing.T) { suite.Run(t, new(CreateUserSuite)) }

func (s *CreateUserSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepo = ifacemocks.NewUserRepository(s.T())
	s.hasher = ifacemocks.NewHasher(s.T())
	s.sut = usecase.NewCreateUser(s.userRepo, s.hasher)
}

func (s *CreateUserSuite) TestExecute() {
	email, _ := vos.NewEmail("new@example.com")
	hashedPwd, _ := vos.NewHashedPassword("hashed123")
	existingUser := entities.Rehydrate(vos.NewUserID(), "Existing", email, hashedPwd, time.Now(), time.Now(), true)

	type args struct{ in dtos.UserRequest }
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(out usecase.CreateUserResult, err error)
	}{
		{
			name: "sucesso criado=true",
			args: args{in: dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(nil, nil).Once()
				s.hasher.EXPECT().Hash("plainpwd").Return(hashedPwd, nil).Once()
				s.userRepo.EXPECT().Add(s.ctx, mock.AnythingOfType("*entities.User")).Return(nil).Once()
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.NoError(err)
				s.True(out.Created)
				s.Equal("new@example.com", out.User.Email)
			},
		},
		{
			name: "idempotente criado=false Add nao chamado",
			args: args{in: dtos.UserRequest{Name: "Existing", Email: "new@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(existingUser, nil).Once()
				s.hasher.EXPECT().Verify(hashedPwd, "plainpwd").Return(true).Once()
				// Add must NOT be called — mock will fail if it is
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.NoError(err)
				s.False(out.Created)
				s.Equal(email.String(), out.User.Email)
			},
		},
		{
			name: "divergencia ErrUserAlreadyExists Add nao chamado",
			args: args{in: dtos.UserRequest{Name: "Existing", Email: "new@example.com", Password: "different"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(existingUser, nil).Once()
				s.hasher.EXPECT().Verify(hashedPwd, "different").Return(false).Once()
				// Add must NOT be called
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.ErrorIs(err, domain.ErrUserAlreadyExists)
			},
		},
		{
			name: "falha de hash",
			args: args{in: dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(nil, nil).Once()
				s.hasher.EXPECT().Hash("plainpwd").Return(vos.HashedPassword(""), errors.New("hash error")).Once()
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.Error(err)
			},
		},
		{
			name: "falha de repo Add",
			args: args{in: dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(nil, nil).Once()
				s.hasher.EXPECT().Hash("plainpwd").Return(hashedPwd, nil).Once()
				s.userRepo.EXPECT().Add(s.ctx, mock.AnythingOfType("*entities.User")).Return(errors.New("db error")).Once()
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.Error(err)
			},
		},
		{
			name: "falha de repo GetByEmail",
			args: args{in: dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"}},
			setup: func() {
				s.userRepo.EXPECT().GetByEmail(s.ctx, email).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out usecase.CreateUserResult, err error) {
				s.Error(err)
			},
		},
		{
			name:  "email invalido",
			args:  args{in: dtos.UserRequest{Name: "New User", Email: "not-valid", Password: "plainpwd"}},
			setup: func() {},
			expect: func(out usecase.CreateUserResult, err error) {
				s.ErrorIs(err, domain.ErrInvalidEmail)
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
