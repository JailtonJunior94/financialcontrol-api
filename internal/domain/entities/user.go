package entities

type User struct {
	Name     string `db:"Name"`
	Email    string `db:"Email"`
	Password string `db:"Password"`
	Entity
}

func NewUser(name, email, password string) *User {
	user := &User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	user.Entity.NewEntity()

	return user
}
