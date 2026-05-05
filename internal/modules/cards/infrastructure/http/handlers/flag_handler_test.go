package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

type FlagHandlerSuite struct {
	suite.Suite
	listFlags *ucmocks.ListFlags
	sut       *handlers.FlagHandler
	app       *fiber.App
	appNoAuth *fiber.App
}

func TestFlagHandlerSuite(t *testing.T) { suite.Run(t, new(FlagHandlerSuite)) }

func (s *FlagHandlerSuite) SetupTest() {
	s.listFlags = ucmocks.NewListFlags(s.T())
	s.sut = handlers.NewFlagHandler(s.listFlags)

	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}

	s.app = fiber.New()
	s.app.Get("/cards/flags", withIdentity, s.sut.List)

	s.appNoAuth = fiber.New()
	s.appNoAuth.Get("/cards/flags", s.sut.List)
}

func (s *FlagHandlerSuite) TestList() {
	flagResp := dtos.FlagResponse{ID: "550e8400-e29b-41d4-a716-446655440010", Name: "Visa", Active: true}
	scenarios := []struct {
		name   string
		path   string
		useApp *fiber.App
		setup  func()
		expect func(status int, body any)
	}{
		{
			name:   "200 lista de bandeiras",
			path:   "/cards/flags",
			useApp: s.app,
			setup: func() {
				s.listFlags.EXPECT().Execute(mock.Anything).Return([]dtos.FlagResponse{flagResp}, nil).Once()
			},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusOK, status)
				arr := body.([]any)
				s.Len(arr, 1)
			},
		},
		{
			name:   "200 query params nao truncam catalogo",
			path:   "/cards/flags?page=1&size=1",
			useApp: s.app,
			setup: func() {
				s.listFlags.EXPECT().Execute(mock.Anything).Return([]dtos.FlagResponse{flagResp}, nil).Once()
			},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusOK, status)
				arr := body.([]any)
				s.Len(arr, 1)
			},
		},
		{
			name:   "401 sem identidade",
			path:   "/cards/flags",
			useApp: s.appNoAuth,
			setup:  func() {},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusUnauthorized, status)
			},
		},
		{
			name:   "500 erro interno",
			path:   "/cards/flags",
			useApp: s.app,
			setup: func() {
				s.listFlags.EXPECT().Execute(mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("GET", sc.path, nil)
			resp, err := sc.useApp.Test(req)
			s.Require().NoError(err)
			var body any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}
