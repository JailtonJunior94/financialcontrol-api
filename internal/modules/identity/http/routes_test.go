package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/pkg/web"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	identityhttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type authServiceStub struct {
	authenticateResponseStatus int
	authenticateResponseData   any
	meResponseStatus           int
	meResponseData             any
}

func (s *authServiceStub) Authenticate(request *identityapp.AuthRequest) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.authenticateResponseStatus, Data: s.authenticateResponseData}
}

func (s *authServiceStub) Me(userID string) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.meResponseStatus, Data: s.meResponseData}
}

type claimsResolverStub struct {
	userID string
	err    error
}

func (s *claimsResolverStub) UserID(authorizationHeader string) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.userID, nil
}

func TestAuthenticateContractPreservesTokenEndpoint(t *testing.T) {
	app := fiber.New()
	controller := identityhttp.NewAuthController(
		&authServiceStub{
			authenticateResponseStatus: fiber.StatusOK,
			authenticateResponseData: fiber.Map{
				"token": "jwt-token",
			},
		},
		&claimsResolverStub{},
	)

	app.Post("/token", controller.Authenticate)

	body, err := json.Marshal(map[string]string{
		"email":    "john@example.com",
		"password": "secret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/token", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMeContractReturnsUnauthorizedWhenClaimsAreInvalid(t *testing.T) {
	app := fiber.New()
	controller := identityhttp.NewAuthController(
		&authServiceStub{
			meResponseStatus: fiber.StatusOK,
			meResponseData:   fiber.Map{"id": "user-id"},
		},
		&claimsResolverStub{err: errors.New("invalid token")},
	)

	app.Get("/me", controller.Me)

	req := httptest.NewRequest(fiber.MethodGet, "/me", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
