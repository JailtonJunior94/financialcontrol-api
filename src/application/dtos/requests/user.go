package requests

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
)

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserRequest) IsValid() error {
	if u.Name == "" {
		return customErrors.NameIsRequired
	}

	if u.Email == "" {
		return customErrors.EmailIsRequired
	}

	if u.Password == "" {
		return customErrors.PasswordIsRequired
	}

	return nil
}
