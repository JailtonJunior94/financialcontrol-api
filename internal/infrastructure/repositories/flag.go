package repositories

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/queries"
)

type FlagRepository struct {
	Db database.ISqlConnection
}

func NewFlagRepository(db database.ISqlConnection) interfaces.IFlagRepository {
	return &FlagRepository{Db: db}
}

func (r *FlagRepository) GetFlags() (flags []entities.Flag, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&flags, queries.GetFlags); err != nil {
		return nil, err
	}
	return flags, nil
}
