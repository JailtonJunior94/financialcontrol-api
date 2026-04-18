package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
)

type IUserService interface {
	CreateUser(request *requests.UserRequest) *responses.HttpResponse
}
