package domain

import "github.com/jailtonjunior94/financialcontrol-api/pkg/entity"

type User struct {
	Name     string `db:"Name"`
	Email    string `db:"Email"`
	Password string `db:"Password"`
	entity.Entity
}

func NewUser(name, email, password string) *User {
	u := &User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	u.NewEntity()
	return u
}
