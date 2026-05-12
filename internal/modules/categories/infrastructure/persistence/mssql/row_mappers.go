package mssql

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

// CategoryRow holds the DB columns for a Category scan.
type CategoryRow struct {
	ID        string    `db:"Id"`
	Name      string    `db:"Name"`
	Sequence  int       `db:"Sequence"`
	CreatedAt time.Time `db:"CreatedAt"`
	UpdatedAt time.Time `db:"UpdatedAt"`
	Active    bool      `db:"Active"`
}

// RowToCategory maps a CategoryRow into the domain Category aggregate.
func RowToCategory(r *CategoryRow) (*entities.Category, error) {
	id, err := vos.ParseCategoryID(r.ID)
	if err != nil {
		return nil, err
	}
	name, err := vos.NewCategoryName(r.Name)
	if err != nil {
		return nil, err
	}

	return entities.RehydrateCategory(
		id,
		name,
		r.Sequence,
		r.CreatedAt.UTC(),
		r.UpdatedAt.UTC(),
		r.Active,
	), nil
}
