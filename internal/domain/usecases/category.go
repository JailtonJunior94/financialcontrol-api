package usecases

import "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"

type ICategoryService interface {
	Categories() *responses.HttpResponse
}
