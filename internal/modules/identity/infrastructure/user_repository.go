package infrastructure

import (
	"database/sql"

	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/queries"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	identitydomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
)

type UserRepository struct {
	db database.ISqlConnection
}

func NewUserRepository(db database.ISqlConnection) identityapp.UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Add(user *identitydomain.User) (*identitydomain.User, error) {
	statement, err := u.db.OpenConnectionAndMountStatement(queries.AddUser)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = statement.Close()
	}()

	result, err := statement.Exec(
		sql.Named("id", user.ID),
		sql.Named("name", user.Name),
		sql.Named("email", user.Email),
		sql.Named("password", user.Password),
		sql.Named("createdAt", user.CreatedAt),
		sql.Named("updatedAt", user.UpdatedAt),
		sql.Named("active", user.Active),
	)
	if err := u.db.ValidateResult(result, err); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserRepository) GetByEmail(email string) (*identitydomain.User, error) {
	connection := u.db.Connect()
	row := connection.QueryRow(queries.GetByEmail, sql.Named("email", email))

	user := new(identitydomain.User)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.Active); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserRepository) GetByID(id string) (*identitydomain.User, error) {
	connection := u.db.Connect()
	row := connection.QueryRow(queries.GetByID, sql.Named("id", id))

	user := new(identitydomain.User)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.Active); err != nil {
		return nil, err
	}

	return user, nil
}
