package repositories

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/queries"
)

type CategoryRepository struct {
	Db database.ISqlConnection
}

func NewCategoryRepository(db database.ISqlConnection) interfaces.ICategoryRepository {
	return &CategoryRepository{Db: db}
}

func (r *CategoryRepository) GetCategories() (categories []entities.Category, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&categories, queries.GetCategories); err != nil {
		return nil, err
	}
	return categories, nil
}
