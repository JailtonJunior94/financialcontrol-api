package requests

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
)

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserRequest) IsValid() error {
	if u.Name == "" {
		return customerrors.NameIsRequired
	}

	if u.Email == "" {
		return customerrors.EmailIsRequired
	}

	if u.Password == "" {
		return customerrors.PasswordIsRequired
	}

	return nil
}
