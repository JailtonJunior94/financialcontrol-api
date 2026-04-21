package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthRequest) IsValid() error {
	if a.Email == "" {
		return customerrors.EmailIsRequired
	}

	if a.Password == "" {
		return customerrors.PasswordIsRequired
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
		return customerrors.EmailIsRequired
	}

	if u.Password == "" {
		return customerrors.PasswordIsRequired
	}

	return nil
}
