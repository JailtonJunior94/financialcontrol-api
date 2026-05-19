package cards_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/routes"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

const (
	smokeUserID = "550e8400-e29b-41d4-a716-446655440001"
	smokeCardID = "550e8400-e29b-41d4-a716-446655440002"
	smokeFlagID = "550e8400-e29b-41d4-a716-446655440003"
)

// buildSmokeApp wires handlers with mocks and returns the app plus all mock instances.
func buildSmokeApp(t *testing.T) (
	app *fiber.App,
	listCards *ucmocks.ListCards,
	getCard *ucmocks.GetCard,
	createCard *ucmocks.CreateCard,
	updateCard *ucmocks.UpdateCard,
	deactivate *ucmocks.DeactivateCard,
	listFlags *ucmocks.ListFlags,
) {
	t.Helper()
	listCards = ucmocks.NewListCards(t)
	getCard = ucmocks.NewGetCard(t)
	createCard = ucmocks.NewCreateCard(t)
	updateCard = ucmocks.NewUpdateCard(t)
	deactivate = ucmocks.NewDeactivateCard(t)
	listFlags = ucmocks.NewListFlags(t)

	cardHandler := handlers.NewCardHandler(listCards, getCard, createCard, updateCard, deactivate)
	flagHandler := handlers.NewFlagHandler(listFlags)

	app = fiber.New()
	routes.RegisterCardRoutes(app, cardHandler, flagHandler, smokeProtected())
	return
}

func smokeProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(context.Background(), identitycontext.Identity{
			UserID: smokeUserID,
			Email:  "smoke@test.com",
		}))
		return c.Next()
	}
}

func smokeCard() dtos.CardResponse {
	return dtos.CardResponse{
		ID:             smokeCardID,
		Name:           "Smoke Card",
		Active:         true,
		ExpirationDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Flag:           dtos.FlagResponse{ID: smokeFlagID, Name: "Visa", Active: true},
	}
}

func TestModule_Smoke_GET_Cards(t *testing.T) {
	app, listCards, _, _, _, _, _ := buildSmokeApp(t)
	listCards.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return([]dtos.CardResponse{smokeCard()}, nil).Once()

	req := httptest.NewRequest("GET", "/cards", nil)
	req.Header.Set("Authorization", "Bearer smoke-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestModule_Smoke_GET_CardFlags(t *testing.T) {
	app, _, _, _, _, _, listFlags := buildSmokeApp(t)
	listFlags.EXPECT().Execute(mock.Anything).Return([]dtos.FlagResponse{{ID: smokeFlagID, Name: "Visa", Active: true}}, nil).Once()

	req := httptest.NewRequest("GET", "/cards/flags", nil)
	req.Header.Set("Authorization", "Bearer smoke-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestModule_Smoke_GET_CardByID(t *testing.T) {
	app, _, getCard, _, _, _, _ := buildSmokeApp(t)
	getCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(smokeCard(), nil).Once()

	req := httptest.NewRequest("GET", "/cards/"+smokeCardID, nil)
	req.Header.Set("Authorization", "Bearer smoke-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestModule_Smoke_POST_Cards(t *testing.T) {
	app, _, _, createCard, _, _, _ := buildSmokeApp(t)
	createCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(smokeCard(), nil).Once()

	body, _ := json.Marshal(map[string]any{
		"flagId":         smokeFlagID,
		"name":           "Smoke Card",
		"number":         "123456789012",
		"description":    "desc",
		"closingDay":     10,
		"expirationDate": "2026-01-15T00:00:00Z",
	})
	req := httptest.NewRequest("POST", "/cards", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer smoke-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestModule_Smoke_PUT_CardByID(t *testing.T) {
	app, _, _, _, updateCard, _, _ := buildSmokeApp(t)
	updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(smokeCard(), nil).Once()

	body, _ := json.Marshal(map[string]any{
		"flagId":         smokeFlagID,
		"name":           "Updated Card",
		"number":         "123456789012",
		"description":    "desc",
		"closingDay":     10,
		"expirationDate": "2026-01-15T00:00:00Z",
	})
	req := httptest.NewRequest("PUT", "/cards/"+smokeCardID, strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer smoke-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestModule_Smoke_DELETE_CardByID(t *testing.T) {
	app, _, _, _, _, deactivate, _ := buildSmokeApp(t)
	deactivate.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	req := httptest.NewRequest("DELETE", "/cards/"+smokeCardID, nil)
	req.Header.Set("Authorization", "Bearer smoke-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}
