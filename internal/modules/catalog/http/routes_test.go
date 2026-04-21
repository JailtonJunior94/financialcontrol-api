package http_test

import (
	"net/http/httptest"
	"testing"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"
	cataloghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type flagServiceStub struct {
	status int
	data   any
}

func (s *flagServiceStub) Flags() *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.status, Data: s.data}
}

type categoryServiceStub struct {
	status int
	data   any
}

func (s *categoryServiceStub) Categories() *appresponses.HttpResponse {
	return &appresponses.HttpResponse{StatusCode: s.status, Data: s.data}
}

func TestFlagRoutesPreserveEndpointContract(t *testing.T) {
	app := fiber.New()
	controller := cataloghttp.NewFlagController(&flagServiceStub{
		status: fiber.StatusOK,
		data: []catalogapp.FlagResponse{
			{ID: "flag-id", Name: "Visa", Active: true},
		},
	})

	cataloghttp.AddFlagRouter(app, controller)

	req := httptest.NewRequest(fiber.MethodGet, "/flags", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCategoryControllerReturnsServicePayload(t *testing.T) {
	app := fiber.New()
	controller := cataloghttp.NewCategoryController(&categoryServiceStub{
		status: fiber.StatusOK,
		data: []catalogapp.CategoryResponse{
			{ID: "category-id", Name: "Moradia", Sequence: 1, Active: true},
		},
	})

	app.Get("/categories", controller.Categories)

	req := httptest.NewRequest(fiber.MethodGet, "/categories", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
