package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
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
		return responses.BadRequest(customErrors.InvalidUserOrPassword)
	}

	if isValid := a.HashAdapter.CheckHash(user.Password, request.Password); !isValid {
		return responses.BadRequest(customErrors.InvalidUserOrPassword)
	}

	token, err := a.JwtAdapter.GenerateTokenJWT(user.ID, user.Email)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(responses.NewAuthResponse(token))
}
