package application

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
)

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewAuthResponse(token string) *AuthResponse {
	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(config.ExpirationAt)),
	}
}

type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}
