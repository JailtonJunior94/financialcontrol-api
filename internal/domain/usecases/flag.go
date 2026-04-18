package usecases

import "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"

type IFlagService interface {
	Flags() *responses.HttpResponse
}
