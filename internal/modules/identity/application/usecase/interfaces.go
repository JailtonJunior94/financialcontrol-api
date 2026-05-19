package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
)

type AuthenticateUser interface {
	Execute(ctx context.Context, in dtos.AuthRequest) (dtos.AuthResponse, error)
}

type CreateUser interface {
	Execute(ctx context.Context, in dtos.UserRequest) (CreateUserResult, error)
}

type GetAuthenticatedUser interface {
	Execute(ctx context.Context) (dtos.MeResponse, error)
}
