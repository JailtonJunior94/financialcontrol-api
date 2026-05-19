package dtos

import (
	"errors"
	"time"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r AuthRequest) Validate() error {
	var errs []error
	if r.Email == "" {
		errs = append(errs, errors.New("email é obrigatório"))
	}
	if r.Password == "" {
		errs = append(errs, errors.New("password é obrigatório"))
	}
	return errors.Join(errs...)
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
