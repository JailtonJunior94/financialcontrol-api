package infrastructure

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
)

type CategoryRepository struct {
	db database.ISqlConnection
}

func NewCategoryRepository(db database.ISqlConnection) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetCategories() ([]entities.Category, error) {
	categories := make([]entities.Category, 0)
	connection := r.db.Connect()
	if err := connection.Select(&categories, getCategories); err != nil {
		return nil, err
	}

	return categories, nil
}
