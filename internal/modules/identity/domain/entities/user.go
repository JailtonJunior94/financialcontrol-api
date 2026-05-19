package entities

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type User struct {
	id        vos.UserID
	name      string
	email     vos.Email
	password  vos.HashedPassword
	createdAt time.Time
	updatedAt time.Time
	active    bool
}

func NewUser(name string, email vos.Email, password vos.HashedPassword) (*User, error) {
	if name == "" {
		return nil, domain.ErrInvalidUser
	}
	return &User{
		id:        vos.NewUserID(),
		name:      name,
		email:     email,
		password:  password,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		active:    true,
	}, nil
}

func RehydrateUser(id vos.UserID, name string, email vos.Email, pwd vos.HashedPassword, createdAt, updatedAt time.Time, active bool) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		password:  pwd,
		createdAt: createdAt,
		updatedAt: updatedAt,
		active:    active,
	}
}

func (u *User) ID() vos.UserID               { return u.id }
func (u *User) Name() string                 { return u.name }
func (u *User) Email() vos.Email             { return u.email }
func (u *User) Password() vos.HashedPassword { return u.password }
func (u *User) CreatedAt() time.Time         { return u.createdAt }
func (u *User) UpdatedAt() time.Time         { return u.updatedAt }
func (u *User) Active() bool                 { return u.active }
