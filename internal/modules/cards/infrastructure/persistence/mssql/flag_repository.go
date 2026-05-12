package mssql

import (
	"context"
	"database/sql"
	"fmt"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	pkgdatabase "github.com/jailtonjunior94/financialcontrol-api/pkg/database"
)

var _ interfaces.FlagRepository = (*FlagRepository)(nil)

type FlagRepository struct {
	db devkitdb.DBTX
}

func NewFlagRepository(db devkitdb.DBTX) *FlagRepository {
	return &FlagRepository{db: db}
}

func (r *FlagRepository) List(ctx context.Context) ([]entities.Flag, error) {
	rows, err := r.db.QueryContext(ctx, listFlags)
	if err != nil {
		return nil, fmt.Errorf("mssql: list flags: %w", err)
	}

	// Column order: Id, Name, Active
	return pkgdatabase.ScanAll[entities.Flag](rows, func(r devkitdb.Rows) (entities.Flag, error) {
		var row FlagRow
		if err := r.Scan(&row.ID, &row.Name, &row.Active); err != nil {
			return entities.Flag{}, fmt.Errorf("mssql: list flags scan: %w", err)
		}
		flag, err := RowToFlag(&row)
		if err != nil {
			return entities.Flag{}, fmt.Errorf("mssql: list flags map: %w", err)
		}
		return flag, nil
	})
}

func (r *FlagRepository) Exists(ctx context.Context, id vos.FlagID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, existsFlag, sql.Named("id", id.String())).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("mssql: exists flag: %w", err)
	}
	return count > 0, nil
}
