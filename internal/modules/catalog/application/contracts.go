package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
)

type FlagRepository interface {
	GetFlags() ([]entities.Flag, error)
}

type CategoryRepository interface {
	GetCategories() ([]entities.Category, error)
}

type FlagService interface {
	Flags() *responses.HttpResponse
}

type CategoryService interface {
	Categories() *responses.HttpResponse
}
