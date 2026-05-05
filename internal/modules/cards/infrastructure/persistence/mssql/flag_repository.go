package mssql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
)

var _ interfaces.FlagRepository = (*FlagRepository)(nil)

type FlagRepository struct {
	db *sqlx.DB
}

func NewFlagRepository(db *sqlx.DB) *FlagRepository {
	return &FlagRepository{db: db}
}

func (r *FlagRepository) List(ctx context.Context) ([]entities.Flag, error) {
	rows, err := r.db.QueryxContext(ctx, listFlags)
	if err != nil {
		return nil, fmt.Errorf("mssql: list flags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	flags := make([]entities.Flag, 0)
	for rows.Next() {
		var row FlagRow
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("mssql: list flags scan: %w", err)
		}
		flag, err := RowToFlag(&row)
		if err != nil {
			return nil, fmt.Errorf("mssql: list flags map: %w", err)
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

func (r *FlagRepository) Exists(ctx context.Context, id vos.FlagID) (bool, error) {
	var count int
	err := r.db.QueryRowxContext(ctx, existsFlag, sql.Named("id", id.String())).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("mssql: exists flag: %w", err)
	}
	return count > 0, nil
}
