package jwt

import (
	"context"
	"errors"
	"slices"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type parser struct {
	cfg Config
}

func NewParser(cfg Config) (Parser, error) {
	cfg = setDefaults(cfg)
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return &parser{cfg: cfg}, nil
}

func (p *parser) Parse(_ context.Context, tokenStr string) (Identity, error) {
	var claims jwtClaims
	token, err := gojwt.ParseWithClaims(tokenStr, &claims, p.keyFunc, gojwt.WithExpirationRequired())
	if err != nil {
		return Identity{}, p.mapError(err)
	}
	if !token.Valid {
		return Identity{}, ErrInvalidToken
	}
	if claims.Subject == "" {
		return Identity{}, ErrInvalidToken
	}
	return Identity{UserID: claims.Subject, Email: claims.Email}, nil
}

func (p *parser) keyFunc(token *gojwt.Token) (any, error) {
	method, ok := token.Method.(*gojwt.SigningMethodHMAC)
	if !ok {
		return nil, ErrAlgorithmMismatch
	}
	if slices.Contains(p.cfg.AllowedAlgorithms, method.Alg()) {
		return p.cfg.Secret, nil
	}
	return nil, ErrAlgorithmMismatch
}

func (p *parser) mapError(err error) error {
	if errors.Is(err, ErrAlgorithmMismatch) {
		return ErrAlgorithmMismatch
	}
	if errors.Is(err, gojwt.ErrTokenExpired) {
		return ErrExpiredToken
	}
	return ErrInvalidToken
}
