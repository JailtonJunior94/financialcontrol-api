package responses

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"
)

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewAuthResponse(token string) *AuthResponse {
	return &AuthResponse{Token: token, ExpiresAt: time.Now().Add(time.Hour * time.Duration(config.ExpirationAt))}
}
