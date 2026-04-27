package jwt_test

import (
	"context"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	"github.com/stretchr/testify/suite"
)

type ParserSuite struct {
	suite.Suite
	ctx    context.Context
	cfg    pkgjwt.Config
	issuer pkgjwt.Issuer
	parser pkgjwt.Parser
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

func (s *ParserSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = pkgjwt.Config{
		Secret:            []byte("super-secret-key-with-at-least-32-bytes!!"),
		AccessTTL:         15 * time.Minute,
		AllowedAlgorithms: []string{"HS256"},
		SigningAlgorithm:  "HS256",
		Issuer:            "financialcontrol",
	}

	var err error
	s.issuer, err = pkgjwt.NewIssuer(s.cfg)
	s.Require().NoError(err)

	s.parser, err = pkgjwt.NewParser(s.cfg)
	s.Require().NoError(err)
}

func (s *ParserSuite) TestNewParser() {
	scenarios := []struct {
		name   string
		cfg    pkgjwt.Config
		expect func(p pkgjwt.Parser, err error)
	}{
		{
			name: "valid config returns parser",
			cfg:  s.cfg,
			expect: func(p pkgjwt.Parser, err error) {
				s.NoError(err)
				s.NotNil(p)
			},
		},
		{
			name: "short secret returns ErrSecretTooShort",
			cfg: pkgjwt.Config{
				Secret: []byte("short"),
			},
			expect: func(p pkgjwt.Parser, err error) {
				s.ErrorIs(err, pkgjwt.ErrSecretTooShort)
				s.Nil(p)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			p, err := pkgjwt.NewParser(sc.cfg)
			sc.expect(p, err)
		})
	}
}

func (s *ParserSuite) TestParse() {
	identity := pkgjwt.Identity{UserID: "user-123", Email: "user@example.com"}

	validToken, _, err := s.issuer.Issue(s.ctx, identity)
	s.Require().NoError(err)

	noneToken := s.buildNoneToken(identity.UserID, identity.Email)
	hs512Token := s.buildHS512Token(identity.UserID, identity.Email)
	expiredToken := s.buildExpiredToken(identity.UserID, identity.Email)
	missingSubToken := s.buildMissingSubToken(identity.Email)

	scenarios := []struct {
		name   string
		token  string
		expect func(id pkgjwt.Identity, err error)
	}{
		{
			name:  "round-trip: valid token returns correct identity",
			token: validToken,
			expect: func(id pkgjwt.Identity, err error) {
				s.NoError(err)
				s.Equal(identity.UserID, id.UserID)
				s.Equal(identity.Email, id.Email)
			},
		},
		{
			name:  "alg:none token is rejected",
			token: noneToken,
			expect: func(id pkgjwt.Identity, err error) {
				s.ErrorIs(err, pkgjwt.ErrAlgorithmMismatch)
			},
		},
		{
			name:  "algorithm outside allow-list (HS512) is rejected",
			token: hs512Token,
			expect: func(id pkgjwt.Identity, err error) {
				s.ErrorIs(err, pkgjwt.ErrAlgorithmMismatch)
			},
		},
		{
			name:  "expired token returns ErrExpiredToken",
			token: expiredToken,
			expect: func(id pkgjwt.Identity, err error) {
				s.ErrorIs(err, pkgjwt.ErrExpiredToken)
			},
		},
		{
			name:  "missing sub claim returns ErrInvalidToken",
			token: missingSubToken,
			expect: func(id pkgjwt.Identity, err error) {
				s.ErrorIs(err, pkgjwt.ErrInvalidToken)
			},
		},
		{
			name:  "malformed token returns ErrInvalidToken",
			token: "not.a.token",
			expect: func(id pkgjwt.Identity, err error) {
				s.ErrorIs(err, pkgjwt.ErrInvalidToken)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			id, err := s.parser.Parse(s.ctx, sc.token)
			sc.expect(id, err)
		})
	}
}

func (s *ParserSuite) buildNoneToken(userID, email string) string {
	claims := gojwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   gojwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodNone, claims)
	signed, err := token.SignedString(gojwt.UnsafeAllowNoneSignatureType)
	s.Require().NoError(err)
	return signed
}

func (s *ParserSuite) buildHS512Token(userID, email string) string {
	claims := gojwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   gojwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS512, claims)
	signed, err := token.SignedString(s.cfg.Secret)
	s.Require().NoError(err)
	return signed
}

func (s *ParserSuite) buildExpiredToken(userID, email string) string {
	claims := gojwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   gojwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cfg.Secret)
	s.Require().NoError(err)
	return signed
}

func (s *ParserSuite) buildMissingSubToken(email string) string {
	claims := gojwt.MapClaims{
		"email": email,
		"exp":   gojwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cfg.Secret)
	s.Require().NoError(err)
	return signed
}
