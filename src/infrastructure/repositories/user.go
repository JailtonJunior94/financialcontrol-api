package repositories

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
)

type UserRepository struct {
	Db database.ISqlConnection
}

func NewUserRepository(db database.ISqlConnection) interfaces.IUserRepository {
	return &UserRepository{Db: db}
}

func (u *UserRepository) Add(p *entities.User) (user *entities.User, err error) {
	return nil, nil
}

func (u *UserRepository) GetByEmail(email string) (user *entities.User, err error) {
	return nil, nil
}
