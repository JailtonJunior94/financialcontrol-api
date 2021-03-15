package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type IAuthService interface {
	Authenticate(request *requests.AuthRequest) *responses.HttpResponse
}
