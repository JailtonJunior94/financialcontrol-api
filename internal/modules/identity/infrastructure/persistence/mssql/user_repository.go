package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

var _ ports.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db devkitdb.DBTX
}

func NewUserRepository(db devkitdb.DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Add(ctx context.Context, u *entities.User) error {
	_, err := r.db.ExecContext(ctx, addUser,
		sql.Named("id", u.ID().String()),
		sql.Named("name", u.Name()),
		sql.Named("email", u.Email().String()),
		sql.Named("password", u.Password().String()),
		sql.Named("createdAt", u.CreatedAt()),
		sql.Named("updatedAt", u.UpdatedAt()),
		sql.Named("active", u.Active()),
	)
	if err != nil {
		return MapDriverError(fmt.Errorf("mssql: add user: %w", err))
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email vos.Email) (*entities.User, error) {
	// Column order: Id, Name, Email, Password, CreatedAt, UpdatedAt, Active
	var rec userRow
	if err := r.db.QueryRowContext(ctx, getUserByEmail, sql.Named("email", email.String())).Scan(
		&rec.ID, &rec.Name, &rec.Email, &rec.Password, &rec.CreatedAt, &rec.UpdatedAt, &rec.Active,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rowToEntity(&rec)
}

func (r *UserRepository) GetByID(ctx context.Context, id vos.UserID) (*entities.User, error) {
	// Column order: Id, Name, Email, Password, CreatedAt, UpdatedAt, Active
	var rec userRow
	if err := r.db.QueryRowContext(ctx, getUserByID, sql.Named("id", id.String())).Scan(
		&rec.ID, &rec.Name, &rec.Email, &rec.Password, &rec.CreatedAt, &rec.UpdatedAt, &rec.Active,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rowToEntity(&rec)
}
