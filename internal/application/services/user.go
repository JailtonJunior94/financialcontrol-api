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

type UserService struct {
	UserRepository interfaces.IUserRepository
	HashAdapter    adapters.IHashAdapter
}

func NewUserService(r interfaces.IUserRepository, h adapters.IHashAdapter) usecases.IUserService {
	return &UserService{UserRepository: r, HashAdapter: h}
}

func (u *UserService) CreateUser(request *requests.UserRequest) *responses.HttpResponse {
	passwordCript, err := u.HashAdapter.GenerateHash(request.Password)
	if err != nil {
		return responses.BadRequest(customerrors.ErrorCreateUserMessage)
	}

	newUser := mappings.ToUserEntity(request, passwordCript)
	user, err := u.UserRepository.Add(newUser)
	if err != nil {
		return responses.ServerError()
	}

	result := mappings.ToUserResponse(user)
	return responses.Created(result)
}
