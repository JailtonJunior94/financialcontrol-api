package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"

	"github.com/gofiber/fiber/v2"
)

type AuthHandlerSuite struct {
	suite.Suite
	authenticateUser     *ucmocks.AuthenticateUser
	getAuthenticatedUser *ucmocks.GetAuthenticatedUser
	sut                  *handlers.AuthHandler
	app                  *fiber.App
}

func TestAuthHandlerSuite(t *testing.T) { suite.Run(t, new(AuthHandlerSuite)) }

func (s *AuthHandlerSuite) SetupTest() {
	s.authenticateUser = ucmocks.NewAuthenticateUser(s.T())
	s.getAuthenticatedUser = ucmocks.NewGetAuthenticatedUser(s.T())
	s.sut = handlers.NewAuthHandler(s.authenticateUser, s.getAuthenticatedUser)

	s.app = fiber.New()
	s.app.Post("/token", s.sut.Authenticate)
	s.app.Get("/me", s.sut.Me)
}

func (s *AuthHandlerSuite) TestAuthenticate() {
	type args struct {
		body string
	}
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(statusCode int, body map[string]any)
	}{
		{
			name: "sucesso",
			args: args{body: `{"email":"user@example.com","password":"secret"}`},
			setup: func() {
				expiresAt := time.Date(2026, time.April, 26, 15, 4, 5, 0, time.UTC)
				s.authenticateUser.EXPECT().
					Execute(mock.Anything, dtos.AuthRequest{Email: "user@example.com", Password: "secret"}).
					Return(dtos.AuthResponse{Token: "tok", ExpiresAt: expiresAt}, nil).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal("tok", body["token"])
				s.Equal("2026-04-26T15:04:05Z", body["expires_at"])
				s.NotContains(body, "access_token")
			},
		},
		{
			name:  "body inválido",
			args:  args{body: `not-json`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:  "email ausente",
			args:  args{body: `{"password":"secret"}`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name:  "senha ausente",
			args:  args{body: `{"email":"user@example.com"}`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name: "credenciais inválidas → 400",
			args: args{body: `{"email":"user@example.com","password":"wrong"}`},
			setup: func() {
				s.authenticateUser.EXPECT().
					Execute(mock.Anything, dtos.AuthRequest{Email: "user@example.com", Password: "wrong"}).
					Return(dtos.AuthResponse{}, domain.ErrInvalidCredentials).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name: "falha de issuer → 500",
			args: args{body: `{"email":"user@example.com","password":"secret"}`},
			setup: func() {
				s.authenticateUser.EXPECT().
					Execute(mock.Anything, dtos.AuthRequest{Email: "user@example.com", Password: "secret"}).
					Return(dtos.AuthResponse{}, domain.ErrTokenIssuance).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("POST", "/token", bytes.NewBufferString(sc.args.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := s.app.Test(req)
			s.Require().NoError(err)

			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

func (s *AuthHandlerSuite) TestMe() {
	scenarios := []struct {
		name   string
		setup  func()
		expect func(statusCode int, body map[string]any)
	}{
		{
			name: "sucesso",
			setup: func() {
				s.getAuthenticatedUser.EXPECT().
					Execute(mock.Anything).
					Return(dtos.MeResponse{ID: "id1", Name: "User", Email: "user@example.com"}, nil).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal("id1", body["id"])
				s.Equal("User", body["name"])
				s.Equal("user@example.com", body["email"])
			},
		},
		{
			name: "identidade ausente → 401",
			setup: func() {
				s.getAuthenticatedUser.EXPECT().
					Execute(mock.Anything).
					Return(dtos.MeResponse{}, identitycontext.ErrNoIdentity).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnauthorized, status)
			},
		},
		{
			name: "usuário não encontrado → 400",
			setup: func() {
				s.getAuthenticatedUser.EXPECT().
					Execute(mock.Anything).
					Return(dtos.MeResponse{}, domain.ErrUserNotFound).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("GET", "/me", nil)
			resp, err := s.app.Test(req)
			s.Require().NoError(err)

			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

// compile-time guard: GetAuthenticatedUser mock satisfies the interface.
var _ usecase.GetAuthenticatedUser = (*ucmocks.GetAuthenticatedUser)(nil)
