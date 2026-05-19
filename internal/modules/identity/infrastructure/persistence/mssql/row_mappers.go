package mssql

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

// userRow holds the DB columns for a User scan. It keeps db tags out of the
// domain entity, which has none (ADR-004).
type userRow struct {
	ID        string    `db:"Id"`
	Name      string    `db:"Name"`
	Email     string    `db:"Email"`
	Password  string    `db:"Password"`
	CreatedAt time.Time `db:"CreatedAt"`
	UpdatedAt time.Time `db:"UpdatedAt"`
	Active    bool      `db:"Active"`
}

// rowToEntity maps a userRow into a domain User entity (no IO).
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
	return entities.RehydrateUser(id, row.Name, email, pwd, row.CreatedAt, row.UpdatedAt, row.Active), nil
}
