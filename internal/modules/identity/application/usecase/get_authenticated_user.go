package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type getAuthenticatedUser struct {
	repo ports.UserRepository
}

func NewGetAuthenticatedUser(repo ports.UserRepository) GetAuthenticatedUser {
	return &getAuthenticatedUser{repo: repo}
}

func (u *getAuthenticatedUser) Execute(ctx context.Context) (dtos.MeResponse, error) {
	identity, err := identitycontext.FromContext(ctx)
	if err != nil {
		return dtos.MeResponse{}, err
	}

	userID, err := vos.ParseUserID(identity.UserID)
	if err != nil {
		return dtos.MeResponse{}, domain.ErrIdentityInvalid
	}

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		return dtos.MeResponse{}, err
	}
	if user == nil {
		return dtos.MeResponse{}, domain.ErrUserNotFound
	}

	return dtos.MeResponse{
		ID:    user.ID().String(),
		Name:  user.Name(),
		Email: user.Email().String(),
	}, nil
}
