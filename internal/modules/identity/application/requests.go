package application

import (
	identitydomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthRequest) IsValid() error {
	if a.Email == "" {
		return identitydomain.EmailIsRequired
	}

	if a.Password == "" {
		return identitydomain.PasswordIsRequired
	}

	return nil
}

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
		return identitydomain.EmailIsRequired
	}

	if u.Password == "" {
		return identitydomain.PasswordIsRequired
	}

	return nil
}
