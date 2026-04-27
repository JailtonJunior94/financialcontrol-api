package jwt

import (
	"context"
	"time"
)

type Identity struct {
	UserID string
	Email  string
}

type Config struct {
	Secret            []byte
	AccessTTL         time.Duration
	AllowedAlgorithms []string
	SigningAlgorithm  string
	Issuer            string
}

type Issuer interface {
	Issue(ctx context.Context, id Identity) (token string, expiresAt time.Time, err error)
}

type Parser interface {
	Parse(ctx context.Context, token string) (Identity, error)
}
