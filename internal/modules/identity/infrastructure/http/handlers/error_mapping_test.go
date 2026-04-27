package handlers_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type ErrorMappingSuite struct {
	suite.Suite
	app *fiber.App
}

func TestErrorMappingSuite(t *testing.T) { suite.Run(t, new(ErrorMappingSuite)) }

func (s *ErrorMappingSuite) SetupTest() {
	s.app = fiber.New()
}

func (s *ErrorMappingSuite) TestMapError() {
	scenarios := []struct {
		name           string
		err            error
		expectedStatus int
		expectedKey    string
	}{
		{
			name:           "ErrInvalidEmail → 400",
			err:            domain.ErrInvalidEmail,
			expectedStatus: fiber.StatusBadRequest,
			expectedKey:    "error",
		},
		{
			name:           "ErrInvalidPassword → 400",
			err:            domain.ErrInvalidPassword,
			expectedStatus: fiber.StatusBadRequest,
			expectedKey:    "error",
		},
		{
			name:           "ErrInvalidCredentials → 400",
			err:            domain.ErrInvalidCredentials,
			expectedStatus: fiber.StatusBadRequest,
			expectedKey:    "error",
		},
		{
			name:           "ErrUserNotFound → 400",
			err:            domain.ErrUserNotFound,
			expectedStatus: fiber.StatusBadRequest,
			expectedKey:    "error",
		},
		{
			name:           "ErrUserAlreadyExists → 409",
			err:            domain.ErrUserAlreadyExists,
			expectedStatus: fiber.StatusConflict,
			expectedKey:    "error",
		},
		{
			name:           "ErrNoIdentity → 401",
			err:            identitycontext.ErrNoIdentity,
			expectedStatus: fiber.StatusUnauthorized,
			expectedKey:    "error",
		},
		{
			name:           "ErrInvalidToken → 401",
			err:            pkgjwt.ErrInvalidToken,
			expectedStatus: fiber.StatusUnauthorized,
			expectedKey:    "error",
		},
		{
			name:           "ErrExpiredToken → 401",
			err:            pkgjwt.ErrExpiredToken,
			expectedStatus: fiber.StatusUnauthorized,
			expectedKey:    "error",
		},
		{
			name:           "ErrAlgorithmMismatch → 401",
			err:            pkgjwt.ErrAlgorithmMismatch,
			expectedStatus: fiber.StatusUnauthorized,
			expectedKey:    "error",
		},
		{
			name:           "ErrTokenIssuance → 500",
			err:            domain.ErrTokenIssuance,
			expectedStatus: fiber.StatusInternalServerError,
			expectedKey:    "error",
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			capturedErr := sc.err
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				return handlers.MapError(c, capturedErr)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			s.Require().NoError(err)
			s.Equal(sc.expectedStatus, resp.StatusCode)

			var body map[string]string
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			s.Contains(body, sc.expectedKey)
			s.NotEmpty(body[sc.expectedKey])
		})
	}
}
