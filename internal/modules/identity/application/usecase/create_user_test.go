package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	ifacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

func TestCreateUser_Execute(t *testing.T) {
	ctx := context.Background()
	email, _ := vos.NewEmail("new@example.com")
	hashedPwd, _ := vos.NewHashedPassword("hashed123")
	existingUser := entities.RehydrateUser(vos.NewUserID(), "Existing", email, hashedPwd, time.Now(), time.Now(), true)

	tests := []struct {
		name   string
		in     dtos.UserRequest
		setup  func(*ifacemocks.UserRepository, *ifacemocks.Hasher)
		assert func(t *testing.T, out usecase.CreateUserResult, err error)
	}{
		{
			name: "sucesso criado=true",
			in:   dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(nil, nil).Once()
				hasher.EXPECT().Hash("plainpwd").Return(hashedPwd, nil).Once()
				repo.EXPECT().Add(ctx, mock.AnythingOfType("*entities.User")).Return(nil).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				require.NoError(t, err)
				assert.True(t, out.Created)
				assert.Equal(t, "new@example.com", out.User.Email)
			},
		},
		{
			name: "idempotente criado=false Add nao chamado",
			in:   dtos.UserRequest{Name: "Existing", Email: "new@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(existingUser, nil).Once()
				hasher.EXPECT().Verify(hashedPwd, "plainpwd").Return(true).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				require.NoError(t, err)
				assert.False(t, out.Created)
				assert.Equal(t, email.String(), out.User.Email)
			},
		},
		{
			name: "divergencia ErrUserAlreadyExists Add nao chamado",
			in:   dtos.UserRequest{Name: "Existing", Email: "new@example.com", Password: "different"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(existingUser, nil).Once()
				hasher.EXPECT().Verify(hashedPwd, "different").Return(false).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
			},
		},
		{
			name: "falha de hash",
			in:   dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(nil, nil).Once()
				hasher.EXPECT().Hash("plainpwd").Return(vos.HashedPassword(""), errors.New("hash error")).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "falha de repo Add",
			in:   dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(nil, nil).Once()
				hasher.EXPECT().Hash("plainpwd").Return(hashedPwd, nil).Once()
				repo.EXPECT().Add(ctx, mock.AnythingOfType("*entities.User")).Return(errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "falha de repo GetByEmail",
			in:   dtos.UserRequest{Name: "New User", Email: "new@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {
				repo.EXPECT().GetByEmail(ctx, email).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:  "email invalido",
			in:    dtos.UserRequest{Name: "New User", Email: "not-valid", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher) {},
			assert: func(t *testing.T, out usecase.CreateUserResult, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidEmail)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ifacemocks.NewUserRepository(t)
			hasher := ifacemocks.NewHasher(t)
			tt.setup(repo, hasher)
			sut := usecase.NewCreateUser(repo, hasher)
			out, err := sut.Execute(ctx, tt.in)
			tt.assert(t, out, err)
		})
	}
}
