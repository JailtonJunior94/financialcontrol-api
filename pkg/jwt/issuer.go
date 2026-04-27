package jwt

import (
	"context"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	Email string `json:"email"`
	gojwt.RegisteredClaims
}

type issuer struct {
	cfg Config
}

func NewIssuer(cfg Config) (Issuer, error) {
	cfg = setDefaults(cfg)
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return &issuer{cfg: cfg}, nil
}

func (i *issuer) Issue(_ context.Context, id Identity) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(i.cfg.AccessTTL)

	claims := jwtClaims{
		Email: id.Email,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   id.UserID,
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
			IssuedAt:  gojwt.NewNumericDate(now),
			Issuer:    i.cfg.Issuer,
		},
	}

	method := gojwt.GetSigningMethod(i.cfg.SigningAlgorithm)
	token := gojwt.NewWithClaims(method, claims)

	signed, err := token.SignedString(i.cfg.Secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}
