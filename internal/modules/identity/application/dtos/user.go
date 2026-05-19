package dtos

import "errors"

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r UserRequest) Validate() error {
	var errs []error
	if r.Name == "" {
		errs = append(errs, errors.New("name é obrigatório"))
	}
	if r.Email == "" {
		errs = append(errs, errors.New("email é obrigatório"))
	}
	if r.Password == "" {
		errs = append(errs, errors.New("password é obrigatório"))
	}
	return errors.Join(errs...)
}

type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

type MeResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
