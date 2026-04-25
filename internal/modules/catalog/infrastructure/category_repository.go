package infrastructure

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
)

type CategoryRepository struct {
	db database.ISqlConnection
}

func NewCategoryRepository(db database.ISqlConnection) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetCategories() ([]domain.Category, error) {
	categories := make([]domain.Category, 0)
	connection := r.db.Connect()
	if err := connection.Select(&categories, getCategories); err != nil {
		return nil, err
	}

	return categories, nil
}
