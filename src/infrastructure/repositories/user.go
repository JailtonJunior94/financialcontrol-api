package repositories

import (
	"database/sql"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/queries"
)

type UserRepository struct {
	Db database.ISqlConnection
}

func NewUserRepository(db database.ISqlConnection) interfaces.IUserRepository {
	return &UserRepository{Db: db}
}

func (u *UserRepository) Add(p *entities.User) (user *entities.User, err error) {
	connection := u.Db.Connect()
	s, err := connection.Prepare(queries.AddUser)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result, err := s.Exec(
		sql.Named("id", p.ID),
		sql.Named("name", p.Name),
		sql.Named("email", p.Email),
		sql.Named("password", p.Password),
		sql.Named("createdAt", p.CreatedAt),
		sql.Named("updatedAt", p.UpdatedAt),
		sql.Named("active", p.Active))

	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return nil, err
	}

	return p, nil
}

func (u *UserRepository) GetByEmail(email string) (user *entities.User, err error) {
	connection := u.Db.Connect()
	row := connection.QueryRow(queries.GetByEmail, sql.Named("email", email))

	usu := new(entities.User)

	if err := row.Scan(&usu.ID, &usu.Name, &usu.Email, &usu.Password, &usu.CreatedAt, &usu.UpdatedAt, &usu.Active); err != nil {
		return nil, err
	}
	return usu, nil
}
