package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
)

type IAuthService interface {
	Authenticate(request *requests.AuthRequest) *responses.HttpResponse
	Me(userID string) *responses.HttpResponse
}
