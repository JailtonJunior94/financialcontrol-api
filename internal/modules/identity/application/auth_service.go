package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
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

func (a *DefaultAuthService) Authenticate(request *AuthRequest) *web.HttpResponse {
	user, err := a.userRepository.GetByEmail(request.Email)
	if err != nil {
		return web.ServerError()
	}

	if user == nil {
		return web.BadRequest(customerrors.InvalidUserOrPassword)
	}

	if isValid := a.hashAdapter.CheckHash(user.Password, request.Password); !isValid {
		return web.BadRequest(customerrors.InvalidUserOrPassword)
	}

	token, err := a.tokenAdapter.GenerateTokenJWT(user.ID, user.Email)
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(NewAuthResponse(token))
}

func (a *DefaultAuthService) Me(userID string) *web.HttpResponse {
	user, err := a.userRepository.GetByID(userID)
	if err != nil {
		return web.ServerError()
	}

	if user == nil {
		return web.BadRequest(customerrors.InvalidUserOrPassword)
	}

	return web.Ok(ToUserResponse(user))
}
