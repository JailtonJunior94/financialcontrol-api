package responses

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"
)

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewAuthResponse(token string) *AuthResponse {
	return &AuthResponse{Token: token, ExpiresAt: time.Now().Local().Add(time.Hour * 24 * time.Duration(environments.ExpirationAt))}
}
