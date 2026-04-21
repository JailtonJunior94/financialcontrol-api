package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type DefaultAuthService struct {
	userRepository UserRepository
	hashAdapter    HashAdapter
	tokenAdapter   TokenAdapter
}

func NewAuthService(userRepository UserRepository, hashAdapter HashAdapter, tokenAdapter TokenAdapter) AuthService {
	return &DefaultAuthService{
		userRepository: userRepository,
		hashAdapter:    hashAdapter,
		tokenAdapter:   tokenAdapter,
	}
}

func (a *DefaultAuthService) Authenticate(request *AuthRequest) *responses.HttpResponse {
	user, err := a.userRepository.GetByEmail(request.Email)
	if err != nil {
		return responses.ServerError()
	}

	if user == nil {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	if isValid := a.hashAdapter.CheckHash(user.Password, request.Password); !isValid {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	token, err := a.tokenAdapter.GenerateTokenJWT(user.ID, user.Email)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(NewAuthResponse(token))
}

func (a *DefaultAuthService) Me(userID string) *responses.HttpResponse {
	user, err := a.userRepository.GetByID(userID)
	if err != nil {
		return responses.ServerError()
	}

	if user == nil {
		return responses.BadRequest(customerrors.InvalidUserOrPassword)
	}

	return responses.Ok(ToUserResponse(user))
}
