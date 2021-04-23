package repositories

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type FlagRepository struct {
	Db database.ISqlConnection
}

func NewFlagRepository(db database.ISqlConnection) interfaces.IFlagRepository {
	return &FlagRepository{Db: db}
}

func (r *FlagRepository) GetFlags() (Flags []entities.Flag, err error) {
	connection := r.Db.Connect()
	if err := connection.Select(&Flags, queries.GetFlags); err != nil {
		return nil, err
	}
	return Flags, nil
}
