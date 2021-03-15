package requests

import "github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthRequest) IsValid() error {
	if a.Email == "" {
		return customErrors.EmailIsRequired
	}

	if a.Password == "" {
		return customErrors.PasswordIsRequired
	}

	return nil
}
