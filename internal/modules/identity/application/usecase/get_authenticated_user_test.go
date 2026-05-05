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
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type GetAuthenticatedUserSuite struct {
	suite.Suite
	ctx      context.Context
	userRepo *ifacemocks.UserRepository
	sut      usecase.GetAuthenticatedUser
}

func TestGetAuthenticatedUserSuite(t *testing.T) { suite.Run(t, new(GetAuthenticatedUserSuite)) }

func (s *GetAuthenticatedUserSuite) SetupTest() {
	s.ctx = context.Background()
	s.userRepo = ifacemocks.NewUserRepository(s.T())
	s.sut = usecase.NewGetAuthenticatedUser(s.userRepo)
}

func (s *GetAuthenticatedUserSuite) TestExecute() {
	userID := vos.NewUserID()
	email, _ := vos.NewEmail("me@example.com")
	pwd, _ := vos.NewHashedPassword("hash")
	user := entities.Rehydrate(userID, "Me User", email, pwd, time.Now(), time.Now(), true)

	ctxWithIdentity := identitycontext.WithIdentity(s.ctx, identitycontext.Identity{
		UserID: userID.String(),
		Email:  email.String(),
	})
	ctxWithInvalidIdentity := identitycontext.WithIdentity(s.ctx, identitycontext.Identity{
		UserID: "not-a-uuid",
		Email:  email.String(),
	})

	scenarios := []struct {
		name   string
		ctx    context.Context
		setup  func()
		expect func(out dtos.MeResponse, err error)
	}{
		{
			name:  "identidade ausente no contexto",
			ctx:   s.ctx,
			setup: func() {},
			expect: func(out dtos.MeResponse, err error) {
				s.ErrorIs(err, identitycontext.ErrNoIdentity)
			},
		},
		{
			name: "identidade presente e repo ok",
			ctx:  ctxWithIdentity,
			setup: func() {
				s.userRepo.EXPECT().GetByID(ctxWithIdentity, userID).Return(user, nil).Once()
			},
			expect: func(out dtos.MeResponse, err error) {
				s.NoError(err)
				s.Equal(userID.String(), out.ID)
				s.Equal("Me User", out.Name)
				s.Equal(email.String(), out.Email)
			},
		},
		{
			name:  "user id invalido no contexto retorna ErrIdentityInvalid (BUG-IDV-003)",
			ctx:   ctxWithInvalidIdentity,
			setup: func() {},
			expect: func(out dtos.MeResponse, err error) {
				s.ErrorIs(err, domain.ErrIdentityInvalid)
				s.NotErrorIs(err, domain.ErrUserNotFound)
			},
		},
		{
			name: "repo erro",
			ctx:  ctxWithIdentity,
			setup: func() {
				s.userRepo.EXPECT().GetByID(ctxWithIdentity, userID).Return(nil, errors.New("db error")).Once()
			},
			expect: func(out dtos.MeResponse, err error) {
				s.Error(err)
			},
		},
		{
			name: "usuario nao encontrado no repo",
			ctx:  ctxWithIdentity,
			setup: func() {
				s.userRepo.EXPECT().GetByID(ctxWithIdentity, userID).Return(nil, nil).Once()
			},
			expect: func(out dtos.MeResponse, err error) {
				s.ErrorIs(err, domain.ErrUserNotFound)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			out, err := s.sut.Execute(sc.ctx)
			sc.expect(out, err)
		})
	}
}
