package requests

import "errors"

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserRequest) IsValid() error {
	if u.Name == "" {
		return errors.New("")
	}

	if u.Email == "" {
		return errors.New("")
	}

	if u.Password == "" {
		return errors.New("")
	}

	return nil
}
