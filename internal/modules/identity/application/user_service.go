package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type DefaultUserService struct {
	userRepository UserRepository
	hashAdapter    HashAdapter
}

func NewUserService(userRepository UserRepository, hashAdapter HashAdapter) UserService {
	return &DefaultUserService{
		userRepository: userRepository,
		hashAdapter:    hashAdapter,
	}
}

func (u *DefaultUserService) CreateUser(request *UserRequest) *web.HttpResponse {
	passwordHash, err := u.hashAdapter.GenerateHash(request.Password)
	if err != nil {
		return web.BadRequest(customerrors.ErrorCreateUserMessage)
	}

	newUser := ToUserEntity(request, passwordHash)
	user, err := u.userRepository.Add(newUser)
	if err != nil {
		return web.ServerError()
	}

	return web.Created(ToUserResponse(user))
}
