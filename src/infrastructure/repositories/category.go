package repositories

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
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
