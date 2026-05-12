package mssql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

// userRow is a private struct for scanning/exec with db tags. It keeps db concerns
// out of the domain entity, which has no tags (ADR-004).
type userRow struct {
	ID        string    `db:"Id"`
	Name      string    `db:"Name"`
	Email     string    `db:"Email"`
	Password  string    `db:"Password"`
	CreatedAt time.Time `db:"CreatedAt"`
	UpdatedAt time.Time `db:"UpdatedAt"`
	Active    bool      `db:"Active"`
}

var _ interfaces.UserRepository = (*UserRepository)(nil)

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
	return err
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

func rowToEntity(row *userRow) (*entities.User, error) {
	id, err := vos.ParseUserID(row.ID)
	if err != nil {
		return nil, err
	}
	email, err := vos.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}
	pwd, err := vos.NewHashedPassword(row.Password)
	if err != nil {
		return nil, err
	}
	return entities.Rehydrate(id, row.Name, email, pwd, row.CreatedAt, row.UpdatedAt, row.Active), nil
}
