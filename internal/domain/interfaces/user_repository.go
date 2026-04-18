package interfaces

import "github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"

type IUserRepository interface {
	Add(p *entities.User) (user *entities.User, err error)
	GetByEmail(email string) (user *entities.User, err error)
	GetByID(id string) (user *entities.User, err error)
}
