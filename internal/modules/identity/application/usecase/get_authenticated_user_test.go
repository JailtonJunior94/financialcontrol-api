package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

func TestGetAuthenticatedUser_Execute(t *testing.T) {
	userID := vos.NewUserID()
	email, _ := vos.NewEmail("me@example.com")
	pwd, _ := vos.NewHashedPassword("hash")
	user := entities.RehydrateUser(userID, "Me User", email, pwd, time.Now(), time.Now(), true)

	baseCtx := context.Background()
	ctxWithIdentity := identitycontext.WithIdentity(baseCtx, identitycontext.Identity{
		UserID: userID.String(),
		Email:  email.String(),
	})
	ctxWithInvalidIdentity := identitycontext.WithIdentity(baseCtx, identitycontext.Identity{
		UserID: "not-a-uuid",
		Email:  email.String(),
	})

	tests := []struct {
		name   string
		ctx    context.Context
		setup  func(*ifacemocks.UserRepository)
		assert func(t *testing.T, out dtos.MeResponse, err error)
	}{
		{
			name:  "identidade ausente no contexto",
			ctx:   baseCtx,
			setup: func(_ *ifacemocks.UserRepository) {},
			assert: func(t *testing.T, out dtos.MeResponse, err error) {
				assert.ErrorIs(t, err, identitycontext.ErrNoIdentity)
			},
		},
		{
			name: "identidade presente e repo ok",
			ctx:  ctxWithIdentity,
			setup: func(repo *ifacemocks.UserRepository) {
				repo.EXPECT().GetByID(ctxWithIdentity, userID).Return(user, nil).Once()
			},
			assert: func(t *testing.T, out dtos.MeResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, userID.String(), out.ID)
				assert.Equal(t, "Me User", out.Name)
				assert.Equal(t, email.String(), out.Email)
			},
		},
		{
			name:  "user id invalido no contexto retorna ErrIdentityInvalid (BUG-IDV-003)",
			ctx:   ctxWithInvalidIdentity,
			setup: func(_ *ifacemocks.UserRepository) {},
			assert: func(t *testing.T, out dtos.MeResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrIdentityInvalid)
				assert.NotErrorIs(t, err, domain.ErrUserNotFound)
			},
		},
		{
			name: "repo erro",
			ctx:  ctxWithIdentity,
			setup: func(repo *ifacemocks.UserRepository) {
				repo.EXPECT().GetByID(ctxWithIdentity, userID).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out dtos.MeResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "usuario nao encontrado no repo",
			ctx:  ctxWithIdentity,
			setup: func(repo *ifacemocks.UserRepository) {
				repo.EXPECT().GetByID(ctxWithIdentity, userID).Return(nil, nil).Once()
			},
			assert: func(t *testing.T, out dtos.MeResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrUserNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ifacemocks.NewUserRepository(t)
			tt.setup(repo)
			sut := usecase.NewGetAuthenticatedUser(repo)
			out, err := sut.Execute(tt.ctx)
			tt.assert(t, out, err)
		})
	}
}
