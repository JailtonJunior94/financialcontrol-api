package handlers_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedKey    string
	}{
		{name: "ErrInvalidEmail → 400", err: domain.ErrInvalidEmail, expectedStatus: fiber.StatusBadRequest, expectedKey: "error"},
		{name: "ErrInvalidPassword → 400", err: domain.ErrInvalidPassword, expectedStatus: fiber.StatusBadRequest, expectedKey: "error"},
		{name: "ErrInvalidCredentials → 400", err: domain.ErrInvalidCredentials, expectedStatus: fiber.StatusBadRequest, expectedKey: "error"},
		{name: "ErrUserNotFound → 400", err: domain.ErrUserNotFound, expectedStatus: fiber.StatusBadRequest, expectedKey: "error"},
		{name: "ErrUserAlreadyExists → 409", err: domain.ErrUserAlreadyExists, expectedStatus: fiber.StatusConflict, expectedKey: "error"},
		{name: "ErrNoIdentity → 401", err: identitycontext.ErrNoIdentity, expectedStatus: fiber.StatusUnauthorized, expectedKey: "error"},
		{name: "ErrIdentityInvalid → 401 (BUG-IDV-003)", err: domain.ErrIdentityInvalid, expectedStatus: fiber.StatusUnauthorized, expectedKey: "error"},
		{name: "ErrInvalidToken → 401", err: pkgjwt.ErrInvalidToken, expectedStatus: fiber.StatusUnauthorized, expectedKey: "error"},
		{name: "ErrExpiredToken → 401", err: pkgjwt.ErrExpiredToken, expectedStatus: fiber.StatusUnauthorized, expectedKey: "error"},
		{name: "ErrAlgorithmMismatch → 401", err: pkgjwt.ErrAlgorithmMismatch, expectedStatus: fiber.StatusUnauthorized, expectedKey: "error"},
		{name: "ErrTokenIssuance → 500", err: domain.ErrTokenIssuance, expectedStatus: fiber.StatusInternalServerError, expectedKey: "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedErr := tt.err
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				status, body := handlers.MapError(capturedErr)
				return c.Status(status).JSON(body)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var body map[string]string
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.Contains(t, body, tt.expectedKey)
			assert.NotEmpty(t, body[tt.expectedKey])
		})
	}
}
