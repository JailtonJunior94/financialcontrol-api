package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
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

func (u *DefaultUserService) CreateUser(request *UserRequest) *responses.HttpResponse {
	passwordHash, err := u.hashAdapter.GenerateHash(request.Password)
	if err != nil {
		return responses.BadRequest(customerrors.ErrorCreateUserMessage)
	}

	newUser := ToUserEntity(request, passwordHash)
	user, err := u.userRepository.Add(newUser)
	if err != nil {
		return responses.ServerError()
	}

	return responses.Created(ToUserResponse(user))
}
