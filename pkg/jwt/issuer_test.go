package jwt_test

import (
	"context"
	"testing"
	"time"

	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	"github.com/stretchr/testify/suite"
)

type IssuerSuite struct {
	suite.Suite
	ctx context.Context
	cfg pkgjwt.Config
}

func TestIssuerSuite(t *testing.T) {
	suite.Run(t, new(IssuerSuite))
}

func (s *IssuerSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = pkgjwt.Config{
		Secret:            []byte("super-secret-key-with-at-least-32-bytes!!"),
		AccessTTL:         15 * time.Minute,
		AllowedAlgorithms: []string{"HS256"},
		SigningAlgorithm:  "HS256",
		Issuer:            "financialcontrol",
	}
}

func (s *IssuerSuite) TestNewIssuer() {
	scenarios := []struct {
		name   string
		cfg    pkgjwt.Config
		expect func(i pkgjwt.Issuer, err error)
	}{
		{
			name: "valid config returns issuer",
			cfg:  s.cfg,
			expect: func(i pkgjwt.Issuer, err error) {
				s.NoError(err)
				s.NotNil(i)
			},
		},
		{
			name: "short secret returns ErrSecretTooShort",
			cfg: pkgjwt.Config{
				Secret:           []byte("short"),
				SigningAlgorithm: "HS256",
			},
			expect: func(i pkgjwt.Issuer, err error) {
				s.ErrorIs(err, pkgjwt.ErrSecretTooShort)
				s.Nil(i)
			},
		},
		{
			name: "signing algorithm not in allowed list returns error",
			cfg: pkgjwt.Config{
				Secret:            []byte("super-secret-key-with-at-least-32-bytes!!"),
				AllowedAlgorithms: []string{"HS256"},
				SigningAlgorithm:  "HS512",
			},
			expect: func(i pkgjwt.Issuer, err error) {
				s.Error(err)
				s.Nil(i)
			},
		},
		{
			name: "defaults applied when signing algorithm and allowed list empty",
			cfg: pkgjwt.Config{
				Secret: []byte("super-secret-key-with-at-least-32-bytes!!"),
			},
			expect: func(i pkgjwt.Issuer, err error) {
				s.NoError(err)
				s.NotNil(i)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			i, err := pkgjwt.NewIssuer(sc.cfg)
			sc.expect(i, err)
		})
	}
}

func (s *IssuerSuite) TestIssue() {
	scenarios := []struct {
		name     string
		identity pkgjwt.Identity
		expect   func(token string, expiresAt time.Time, err error)
	}{
		{
			name:     "valid identity produces non-empty token",
			identity: pkgjwt.Identity{UserID: "user-123", Email: "user@example.com"},
			expect: func(token string, expiresAt time.Time, err error) {
				s.NoError(err)
				s.NotEmpty(token)
				s.True(expiresAt.After(time.Now()))
			},
		},
		{
			name:     "empty userID still produces token",
			identity: pkgjwt.Identity{UserID: "", Email: "user@example.com"},
			expect: func(token string, expiresAt time.Time, err error) {
				s.NoError(err)
				s.NotEmpty(token)
			},
		},
	}

	sut, err := pkgjwt.NewIssuer(s.cfg)
	s.Require().NoError(err)

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			token, expiresAt, err := sut.Issue(s.ctx, sc.identity)
			sc.expect(token, expiresAt, err)
		})
	}
}
