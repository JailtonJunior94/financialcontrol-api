package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

const (
	testUserID = "550e8400-e29b-41d4-a716-446655440001"
	testCardID = "550e8400-e29b-41d4-a716-446655440002"
)

type CardHandlerSuite struct {
	suite.Suite
	listCards  *ucmocks.ListCards
	getCard    *ucmocks.GetCard
	createCard *ucmocks.CreateCard
	updateCard *ucmocks.UpdateCard
	deactivate *ucmocks.DeactivateCard
	sut        *handlers.CardHandler
	app        *fiber.App
}

func TestCardHandlerSuite(t *testing.T) { suite.Run(t, new(CardHandlerSuite)) }

func (s *CardHandlerSuite) SetupTest() {
	s.listCards = ucmocks.NewListCards(s.T())
	s.getCard = ucmocks.NewGetCard(s.T())
	s.createCard = ucmocks.NewCreateCard(s.T())
	s.updateCard = ucmocks.NewUpdateCard(s.T())
	s.deactivate = ucmocks.NewDeactivateCard(s.T())
	s.sut = handlers.NewCardHandler(s.listCards, s.getCard, s.createCard, s.updateCard, s.deactivate)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Get("/cards", withIdentity, s.sut.List)
	s.app.Get("/cards/:id", withIdentity, s.sut.Get)
	s.app.Post("/cards", withIdentity, s.sut.Create)
	s.app.Put("/cards/:id", withIdentity, s.sut.Update)
	s.app.Delete("/cards/:id", withIdentity, s.sut.Deactivate)
}

func (s *CardHandlerSuite) TestList() {
	cardResp := dtos.CardResponse{ID: testCardID, Name: "Meu Cartão", Active: true}
	scenarios := []struct {
		name   string
		path   string
		setup  func()
		expect func(status int, body any)
	}{
		{
			name: "200 lista de cartões sem paginacao por default",
			path: "/cards",
			setup: func() {
				s.listCards.EXPECT().Execute(mock.Anything, mock.Anything, dtos.Pagination{}).Return([]dtos.CardResponse{cardResp}, nil).Once()
			},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusOK, status)
				arr := body.([]any)
				s.Len(arr, 1)
			},
		},
		{
			name: "200 lista de cartões com paginacao normalizada",
			path: "/cards?page=1&size=1000",
			setup: func() {
				s.listCards.EXPECT().Execute(mock.Anything, mock.Anything, dtos.NewPagination(1, 1000)).Return([]dtos.CardResponse{cardResp}, nil).Once()
			},
			expect: func(status int, body any) {
				s.Equal(fiber.StatusOK, status)
				arr := body.([]any)
				s.Len(arr, 1)
			},
		},
		{
			name: "500 erro de repositório",
			path: "/cards",
			setup: func() {
				s.listCards.EXPECT().Execute(mock.Anything, mock.Anything, dtos.Pagination{}).Return(nil, errors.New("db error")).Once()
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
			resp, err := s.app.Test(req)
			s.Require().NoError(err)
			var body any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

func (s *CardHandlerSuite) TestGet() {
	cardResp := dtos.CardResponse{ID: testCardID, Name: "Meu Cartão", Active: true}
	scenarios := []struct {
		name   string
		id     string
		setup  func()
		expect func(status int, body map[string]any)
	}{
		{
			name: "200 cartão encontrado",
			id:   testCardID,
			setup: func() {
				s.getCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(cardResp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal(testCardID, body["id"])
			},
		},
		{
			name:  "400 id inválido",
			id:    "not-a-uuid",
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
				s.Contains(body, "error")
			},
		},
		{
			name: "404 cartão não encontrado",
			id:   testCardID,
			setup: func() {
				s.getCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, domain.ErrCardNotFound).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
		{
			name: "500 erro interno",
			id:   testCardID,
			setup: func() {
				s.getCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, errors.New("db error")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("GET", "/cards/"+sc.id, nil)
			resp, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

func (s *CardHandlerSuite) TestCreate() {
	cardResp := dtos.CardResponse{ID: testCardID, Name: "Meu Cartão", Active: true}
	validBody := `{"flagId":"550e8400-e29b-41d4-a716-446655440000","name":"Meu Cartão","number":"4111111111111111","description":"desc","closingDay":15,"expirationDate":"2026-05-10T00:00:00Z"}`

	scenarios := []struct {
		name        string
		body        string
		contentType string
		setup       func()
		expect      func(status int, body map[string]any)
	}{
		{
			name:        "201 criado com sucesso",
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.createCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(cardResp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusCreated, status)
				s.Equal(testCardID, body["id"])
			},
		},
		{
			name:        "422 body inválido",
			body:        "not-json",
			contentType: "application/json",
			setup:       func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:        "422 validação de campos",
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.createCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, domain.ErrInvalidCardName).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:        "500 erro interno",
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.createCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, errors.New("db error")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("POST", "/cards", bytes.NewBufferString(sc.body))
			req.Header.Set("Content-Type", sc.contentType)
			resp, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

func (s *CardHandlerSuite) TestCreate_ResponseContractPreservesFlagPayload() {
	cardResp := dtos.CardResponse{
		ID:             testCardID,
		Name:           "Meu Cartão",
		Number:         "4111111111111111",
		Description:    "desc",
		ClosingDay:     15,
		BestDayToBuy:   14,
		ExpirationDate: time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
		Active:         true,
		Flag: dtos.FlagResponse{
			ID:     "550e8400-e29b-41d4-a716-446655440000",
			Name:   "Visa",
			Active: true,
		},
	}
	body := `{"flagId":"550e8400-e29b-41d4-a716-446655440000","name":"Meu Cartão","number":"4111111111111111","description":"desc","closingDay":15,"expirationDate":"2026-05-10T00:00:00Z"}`
	s.createCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(cardResp, nil).Once()

	req := httptest.NewRequest("POST", "/cards", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.app.Test(req)
	s.Require().NoError(err)

	rawBody := new(bytes.Buffer)
	_, err = rawBody.ReadFrom(resp.Body)
	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, resp.StatusCode)
	s.Equal(`{"id":"550e8400-e29b-41d4-a716-446655440002","name":"Meu Cartão","number":"4111111111111111","description":"desc","closingDay":15,"bestDayToBuy":14,"expirationDate":"2026-05-10T00:00:00Z","active":true,"flag":{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Visa","active":true}}`, rawBody.String())
}

func (s *CardHandlerSuite) TestGet_WrappedNotFoundUsesCanonicalMessage() {
	s.getCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, fmt.Errorf("usecase get_card: %w", domain.ErrCardNotFound)).Once()

	req := httptest.NewRequest("GET", "/cards/"+testCardID, nil)
	resp, err := s.app.Test(req)
	s.Require().NoError(err)

	rawBody := new(bytes.Buffer)
	_, err = rawBody.ReadFrom(resp.Body)
	s.Require().NoError(err)
	s.Equal(fiber.StatusNotFound, resp.StatusCode)
	s.Equal(`{"error":"Cartão não encontrado"}`, rawBody.String())
}

func (s *CardHandlerSuite) TestUpdate() {
	cardResp := dtos.CardResponse{ID: testCardID, Name: "Meu Cartão", Active: true}
	validBody := `{"flagId":"550e8400-e29b-41d4-a716-446655440000","name":"Meu Cartão","number":"4111111111111111","description":"desc","closingDay":15,"expirationDate":"2026-05-10T00:00:00Z"}`

	scenarios := []struct {
		name        string
		id          string
		body        string
		contentType string
		setup       func()
		expect      func(status int, body map[string]any)
	}{
		{
			name:        "200 atualizado com sucesso",
			id:          testCardID,
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cardResp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal(testCardID, body["id"])
			},
		},
		{
			name:        "400 id inválido",
			id:          "not-a-uuid",
			body:        validBody,
			contentType: "application/json",
			setup:       func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name:        "422 body inválido",
			id:          testCardID,
			body:        "not-json",
			contentType: "application/json",
			setup:       func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:        "422 validação de campos",
			id:          testCardID,
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, domain.ErrInvalidCardName).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:        "404 cartão não encontrado",
			id:          testCardID,
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, domain.ErrCardNotFound).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
		{
			name:        "500 erro interno",
			id:          testCardID,
			body:        validBody,
			contentType: "application/json",
			setup: func() {
				s.updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(dtos.CardResponse{}, errors.New("db error")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("PUT", "/cards/"+sc.id, bytes.NewBufferString(sc.body))
			req.Header.Set("Content-Type", sc.contentType)
			resp, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

func (s *CardHandlerSuite) TestUpdate_ResponseContractPreservesFlagPayload() {
	cardResp := dtos.CardResponse{
		ID:             testCardID,
		Name:           "Meu Cartão Atualizado",
		Number:         "4111111111111111",
		Description:    "desc",
		ClosingDay:     10,
		BestDayToBuy:   9,
		ExpirationDate: time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
		Active:         true,
		Flag: dtos.FlagResponse{
			ID:     "550e8400-e29b-41d4-a716-446655440000",
			Name:   "Visa",
			Active: true,
		},
	}
	body := `{"flagId":"550e8400-e29b-41d4-a716-446655440000","name":"Meu Cartão Atualizado","number":"4111111111111111","description":"desc","closingDay":10,"expirationDate":"2026-05-10T00:00:00Z"}`
	s.updateCard.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cardResp, nil).Once()

	req := httptest.NewRequest("PUT", "/cards/"+testCardID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.app.Test(req)
	s.Require().NoError(err)

	rawBody := new(bytes.Buffer)
	_, err = rawBody.ReadFrom(resp.Body)
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)
	s.Equal(`{"id":"550e8400-e29b-41d4-a716-446655440002","name":"Meu Cartão Atualizado","number":"4111111111111111","description":"desc","closingDay":10,"bestDayToBuy":9,"expirationDate":"2026-05-10T00:00:00Z","active":true,"flag":{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Visa","active":true}}`, rawBody.String())
}

func (s *CardHandlerSuite) TestDeactivate() {
	scenarios := []struct {
		name   string
		id     string
		setup  func()
		expect func(status int)
	}{
		{
			name: "204 desativado com sucesso",
			id:   testCardID,
			setup: func() {
				s.deactivate.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(status int) {
				s.Equal(fiber.StatusNoContent, status)
			},
		},
		{
			name:  "400 id inválido",
			id:    "not-a-uuid",
			setup: func() {},
			expect: func(status int) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name: "404 cartão não encontrado",
			id:   testCardID,
			setup: func() {
				s.deactivate.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(domain.ErrCardNotFound).Once()
			},
			expect: func(status int) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
		{
			name: "500 erro interno",
			id:   testCardID,
			setup: func() {
				s.deactivate.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			expect: func(status int) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("DELETE", "/cards/"+sc.id, nil)
			resp, err := s.app.Test(req)
			s.Require().NoError(err)
			sc.expect(resp.StatusCode)
		})
	}
}
