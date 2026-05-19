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
)

func TestAuthenticateUser_Execute(t *testing.T) {
	ctx := context.Background()
	email, _ := vos.NewEmail("user@example.com")
	pwd, _ := vos.NewHashedPassword("hashedpwd")
	user := entities.RehydrateUser(vos.NewUserID(), "Test User", email, pwd, time.Now(), time.Now(), true)
	expiresAt := time.Now().Add(15 * time.Minute)

	tests := []struct {
		name   string
		in     dtos.AuthRequest
		setup  func(*ifacemocks.UserRepository, *ifacemocks.Hasher, *ifacemocks.TokenIssuer)
		assert func(t *testing.T, out dtos.AuthResponse, err error)
	}{
		{
			name: "sucesso",
			in:   dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher, issuer *ifacemocks.TokenIssuer) {
				repo.EXPECT().GetByEmail(ctx, email).Return(user, nil).Once()
				hasher.EXPECT().Verify(pwd, "plainpwd").Return(true).Once()
				issuer.EXPECT().Issue(ctx, user.ID(), user.Email()).Return("token123", expiresAt, nil).Once()
			},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, "token123", out.Token)
				assert.Equal(t, expiresAt, out.ExpiresAt)
			},
		},
		{
			name:  "e-mail invalido",
			in:    dtos.AuthRequest{Email: "not-an-email", Password: "plainpwd"},
			setup: func(_ *ifacemocks.UserRepository, _ *ifacemocks.Hasher, _ *ifacemocks.TokenIssuer) {},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "e-mail desconhecido",
			in:   dtos.AuthRequest{Email: "unknown@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, _ *ifacemocks.Hasher, _ *ifacemocks.TokenIssuer) {
				unknownEmail, _ := vos.NewEmail("unknown@example.com")
				repo.EXPECT().GetByEmail(ctx, unknownEmail).Return(nil, nil).Once()
			},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "senha incorreta",
			in:   dtos.AuthRequest{Email: "user@example.com", Password: "wrong"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher, _ *ifacemocks.TokenIssuer) {
				repo.EXPECT().GetByEmail(ctx, email).Return(user, nil).Once()
				hasher.EXPECT().Verify(pwd, "wrong").Return(false).Once()
			},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
			},
		},
		{
			name: "falha de repo",
			in:   dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, _ *ifacemocks.Hasher, _ *ifacemocks.TokenIssuer) {
				repo.EXPECT().GetByEmail(ctx, email).Return(nil, errors.New("db error")).Once()
			},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "falha de issuer",
			in:   dtos.AuthRequest{Email: "user@example.com", Password: "plainpwd"},
			setup: func(repo *ifacemocks.UserRepository, hasher *ifacemocks.Hasher, issuer *ifacemocks.TokenIssuer) {
				repo.EXPECT().GetByEmail(ctx, email).Return(user, nil).Once()
				hasher.EXPECT().Verify(pwd, "plainpwd").Return(true).Once()
				issuer.EXPECT().Issue(ctx, user.ID(), user.Email()).Return("", time.Time{}, errors.New("issuer error")).Once()
			},
			assert: func(t *testing.T, out dtos.AuthResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrTokenIssuance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ifacemocks.NewUserRepository(t)
			hasher := ifacemocks.NewHasher(t)
			issuer := ifacemocks.NewTokenIssuer(t)
			tt.setup(repo, hasher, issuer)
			sut := usecase.NewAuthenticateUser(repo, hasher, issuer)
			out, err := sut.Execute(ctx, tt.in)
			tt.assert(t, out, err)
		})
	}
}
