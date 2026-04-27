package usecase

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type CreateUserResult struct {
	User    dtos.UserResponse
	Created bool
}

type CreateUser interface {
	Execute(ctx context.Context, in dtos.UserRequest) (CreateUserResult, error)
}

type createUser struct {
	repo   interfaces.UserRepository
	hasher interfaces.Hasher
}

func NewCreateUser(repo interfaces.UserRepository, hasher interfaces.Hasher) CreateUser {
	return &createUser{repo: repo, hasher: hasher}
}

func (u *createUser) Execute(ctx context.Context, in dtos.UserRequest) (CreateUserResult, error) {
	email, err := vos.NewEmail(in.Email)
	if err != nil {
		return CreateUserResult{}, domain.ErrInvalidEmail
	}

	existing, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return CreateUserResult{}, err
	}

	if existing != nil {
		if u.hasher.Verify(existing.Password(), in.Password) {
			return CreateUserResult{User: toUserResponse(existing), Created: false}, nil
		}
		return CreateUserResult{}, domain.ErrUserAlreadyExists
	}

	hashed, err := u.hasher.Hash(in.Password)
	if err != nil {
		return CreateUserResult{}, err
	}

	user, err := entities.NewUser(in.Name, email, hashed)
	if err != nil {
		return CreateUserResult{}, err
	}

	if err := u.repo.Add(ctx, user); err != nil {
		return CreateUserResult{}, err
	}

	return CreateUserResult{User: toUserResponse(user), Created: true}, nil
}

func toUserResponse(u *entities.User) dtos.UserResponse {
	return dtos.UserResponse{
		ID:     u.ID().String(),
		Name:   u.Name(),
		Email:  u.Email().String(),
		Active: u.Active(),
	}
}
