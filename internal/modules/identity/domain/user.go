package domain

import "github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"

type User = entities.User

func NewUser(name, email, password string) *User {
	return entities.NewUser(name, email, password)
}
