package interfaces

import "github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"

type ICategoryRepository interface {
	GetCategories() (categories []entities.Category, err error)
}
