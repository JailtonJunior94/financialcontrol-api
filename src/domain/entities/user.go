package entities

type User struct {
	Entity
	Name     string `db:"Name"`
	Email    string `db:"Email"`
	Password string `db:"Password"`
}

func (u *User) NewUser(name, email, password string) {
	u.Entity.NewEntity()
	u.Name = name
	u.Email = email
	u.Password = password
}
