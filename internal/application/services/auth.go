package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"
)

type AuthService struct {
	UserRepository interfaces.IUserRepository
	HashAdapter    adapters.IHashAdapter
	JwtAdapter     adapters.IJwtAdapter
}

func NewAuthService(r interfaces.IUserRepository, h adapters.IHashAdapter, j adapters.IJwtAdapter) usecases.IAuthService {
	return &AuthService{UserRepository: r, HashAdapter: h, JwtAdapter: j}
}

func (a *AuthService) Authenticate(request *requests.AuthRequest) *responses.HttpResponse {
	user, err := a.UserRepository.GetByEmail(request.Email)
	if err != nil {
		return responses.ServerError()
	}

	if user == nil {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	if isValid := a.HashAdapter.CheckHash(user.Password, request.Password); !isValid {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	token, err := a.JwtAdapter.GenerateTokenJWT(user.ID, user.Email)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(responses.NewAuthResponse(token))
}

func (a *AuthService) Me(userID string) *responses.HttpResponse {
	user, err := a.UserRepository.GetByID(userID)
	if err != nil {
		return responses.ServerError()
	}

	if user == nil {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	return responses.Ok(mappings.ToUserResponse(user))
}
