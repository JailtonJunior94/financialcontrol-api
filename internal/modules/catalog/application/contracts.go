package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"
)

type FlagRepository interface {
	GetFlags() ([]entities.Flag, error)
}

type CategoryRepository interface {
	GetCategories() ([]entities.Category, error)
}

type FlagService interface {
	Flags() *web.HttpResponse
}

type CategoryService interface {
	Categories() *web.HttpResponse
}
