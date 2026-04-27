package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type AuthenticateUser interface {
	Execute(ctx context.Context, in dtos.AuthRequest) (dtos.AuthResponse, error)
}

type authenticateUser struct {
	repo   interfaces.UserRepository
	hasher interfaces.Hasher
	issuer interfaces.TokenIssuer
}

func NewAuthenticateUser(repo interfaces.UserRepository, hasher interfaces.Hasher, issuer interfaces.TokenIssuer) AuthenticateUser {
	return &authenticateUser{repo: repo, hasher: hasher, issuer: issuer}
}

func (u *authenticateUser) Execute(ctx context.Context, in dtos.AuthRequest) (dtos.AuthResponse, error) {
	email, err := vos.NewEmail(in.Email)
	if err != nil {
		return dtos.AuthResponse{}, domain.ErrInvalidCredentials
	}

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return dtos.AuthResponse{}, err
	}
	if user == nil {
		return dtos.AuthResponse{}, domain.ErrInvalidCredentials
	}

	if !u.hasher.Verify(user.Password(), in.Password) {
		return dtos.AuthResponse{}, domain.ErrInvalidCredentials
	}

	token, expiresAt, err := u.issuer.Issue(ctx, user.ID(), user.Email())
	if err != nil {
		return dtos.AuthResponse{}, domain.ErrTokenIssuance
	}

	return dtos.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
