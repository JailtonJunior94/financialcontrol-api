package entities

type User struct {
	Name     string `db:"Name"`
	Email    string `db:"Email"`
	Password string `db:"Password"`
	Entity
}

func (u *User) NewUser(name, email, password string) {
	u.Entity.NewEntity()
	u.Name = name
	u.Email = email
	u.Password = password
}
