package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/web"
)

type FlagRepository interface {
	GetFlags() ([]domain.Flag, error)
}

type CategoryRepository interface {
	GetCategories() ([]domain.Category, error)
}

type FlagService interface {
	Flags() *web.HttpResponse
}

type CategoryService interface {
	Categories() *web.HttpResponse
}
