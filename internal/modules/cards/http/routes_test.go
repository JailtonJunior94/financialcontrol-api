package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	cardsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application"
	cardshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/http"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	appresponses "github.com/jailtonjunior94/financialcontrol-api/pkg/web"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type cardServiceStub struct {
	cardsStatus    int
	cardsData      any
	cardByIDStatus int
	cardByIDData   any
	createStatus   int
	createData     any
	updateStatus   int
	updateData     any
	removeStatus   int
	removeData     any
}

func (s *cardServiceStub) Cards(userID string) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.cardsStatus, Data: s.cardsData}
}

func (s *cardServiceStub) CardById(id, userID string) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.cardByIDStatus, Data: s.cardByIDData}
}

func (s *cardServiceStub) CreateCard(userID string, request *cardsapp.CardRequest) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.createStatus, Data: s.createData}
}

func (s *cardServiceStub) UpdateCard(id, userID string, request *cardsapp.CardRequest) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.updateStatus, Data: s.updateData}
}

func (s *cardServiceStub) RemoveCard(id, userID string) *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.removeStatus, Data: s.removeData}
}

func withIdentity(app *fiber.App, userID string) *fiber.App {
	app.Use(func(c *fiber.Ctx) error {
		ctx := identitycontext.WithIdentity(context.Background(), identitycontext.Identity{UserID: userID})
		c.SetUserContext(ctx)
		return c.Next()
	})
	return app
}

func TestCardsContractReturnsUnauthorizedWhenIdentityMissing(t *testing.T) {
	app := fiber.New()
	controller := cardshttp.NewCardController(
		&cardServiceStub{cardsStatus: fiber.StatusOK, cardsData: []cardsapp.CardResponse{}},
	)
	app.Get("/cards", controller.Cards)

	req := httptest.NewRequest(fiber.MethodGet, "/cards", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateCardContractPreservesPayloadValidation(t *testing.T) {
	app := fiber.New()
	controller := cardshttp.NewCardController(
		&cardServiceStub{
			createStatus: fiber.StatusCreated,
			createData:   &cardsapp.CardResponse{ID: "card-id", Name: "Cartao XP"},
		},
	)
	withIdentity(app, "user-id")
	app.Post("/cards", controller.CreateCard)

	body, err := json.Marshal(map[string]any{
		"flagId":         "flag-id",
		"name":           "Cartao XP",
		"number":         "1234",
		"description":    "Principal",
		"closingDay":     10,
		"expirationDate": "2027-01-01T00:00:00Z",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/cards", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}
