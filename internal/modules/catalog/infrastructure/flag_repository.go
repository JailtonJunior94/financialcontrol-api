package infrastructure

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
)

type FlagRepository struct {
	db database.ISqlConnection
}

func NewFlagRepository(db database.ISqlConnection) *FlagRepository {
	return &FlagRepository{db: db}
}

func (r *FlagRepository) GetFlags() ([]domain.Flag, error) {
	flags := make([]domain.Flag, 0)
	connection := r.db.Connect()
	if err := connection.Select(&flags, getFlags); err != nil {
		return nil, err
	}

	return flags, nil
}
