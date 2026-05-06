package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
)

const (
	testUserID     = "550e8400-e29b-41d4-a716-446655440001"
	testCategoryID = "550e8400-e29b-41d4-a716-446655440002"
)

type CategoryHandlerSuite struct {
	suite.Suite
	listC   *ucmocks.ListCategories
	getC    *ucmocks.GetCategory
	createC *ucmocks.CreateCategory
	updateC *ucmocks.UpdateCategory
	deleteC *ucmocks.DeleteCategory
	sut     *handlers.CategoryHandler
	app     *fiber.App
}

func TestCategoryHandlerSuite(t *testing.T) { suite.Run(t, new(CategoryHandlerSuite)) }

func (s *CategoryHandlerSuite) SetupTest() {
	s.listC = ucmocks.NewListCategories(s.T())
	s.getC = ucmocks.NewGetCategory(s.T())
	s.createC = ucmocks.NewCreateCategory(s.T())
	s.updateC = ucmocks.NewUpdateCategory(s.T())
	s.deleteC = ucmocks.NewDeleteCategory(s.T())
	s.sut = handlers.NewCategoryHandler(s.listC, s.getC, s.createC, s.updateC, s.deleteC)

	s.app = fiber.New()
	withIdentity := func(c *fiber.Ctx) error {
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: testUserID,
			Email:  "user@example.com",
		}))
		return c.Next()
	}
	s.app.Get("/categories", withIdentity, s.sut.List)
	s.app.Get("/categories/:id", withIdentity, s.sut.Get)
	s.app.Post("/categories", withIdentity, s.sut.Create)
	s.app.Put("/categories/:id", withIdentity, s.sut.Update)
	s.app.Delete("/categories/:id", withIdentity, s.sut.Delete)
}

func sampleResponse() dtos.CategoryResponse {
	return dtos.CategoryResponse{
		ID:        testCategoryID,
		Name:      "Mercado",
		Color:     "blue",
		Icon:      "cart",
		CreatedAt: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.May, 5, 0, 0, 0, 0, time.UTC),
	}
}

func (s *CategoryHandlerSuite) TestList() {
	resp := sampleResponse()
	page := dtos.PaginatedResponse[dtos.CategoryResponse]{
		Items: []dtos.CategoryResponse{resp}, Total: 1, Page: 1, PageSize: 10,
	}
	scenarios := []struct {
		name   string
		path   string
		setup  func()
		expect func(status int, body map[string]any)
	}{
		{
			name: "200 default pagination",
			path: "/categories",
			setup: func() {
				s.listC.EXPECT().Execute(mock.Anything, mock.Anything, dtos.ListCategoriesQuery{}).Return(page, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.EqualValues(1, body["total"])
			},
		},
		{
			name: "200 with filters",
			path: "/categories?page=2&pageSize=10&name=mer&scope=roots",
			setup: func() {
				s.listC.EXPECT().Execute(mock.Anything, mock.Anything, dtos.ListCategoriesQuery{
					Page: 2, PageSize: 10, Name: "mer", Scope: "roots",
				}).Return(page, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
			},
		},
		{
			name:  "422 invalid page",
			path:  "/categories?page=abc",
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "500 use case error",
			path: "/categories",
			setup: func() {
				s.listC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.PaginatedResponse[dtos.CategoryResponse]{}, errors.New("db")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("GET", sc.path, nil)
			res, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			_ = json.NewDecoder(res.Body).Decode(&body)
			sc.expect(res.StatusCode, body)
		})
	}
}

func (s *CategoryHandlerSuite) TestGet() {
	resp := sampleResponse()
	scenarios := []struct {
		name   string
		id     string
		setup  func()
		expect func(status int, body map[string]any)
	}{
		{
			name: "200 found",
			id:   testCategoryID,
			setup: func() {
				s.getC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(resp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal(testCategoryID, body["id"])
			},
		},
		{
			name: "422 invalid id",
			id:   "not-uuid",
			setup: func() {
				s.getC.EXPECT().Execute(mock.Anything, mock.Anything, "not-uuid").Return(dtos.CategoryResponse{}, domain.ErrInvalidCategoryID).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "404 not found",
			id:   testCategoryID,
			setup: func() {
				s.getC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(dtos.CategoryResponse{}, domain.ErrCategoryNotFound).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
		{
			name: "500 internal",
			id:   testCategoryID,
			setup: func() {
				s.getC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(dtos.CategoryResponse{}, errors.New("db")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("GET", "/categories/"+sc.id, nil)
			res, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			_ = json.NewDecoder(res.Body).Decode(&body)
			sc.expect(res.StatusCode, body)
		})
	}
}

func (s *CategoryHandlerSuite) TestCreate() {
	resp := sampleResponse()
	validBody := `{"name":"Mercado","color":"blue","icon":"cart"}`

	scenarios := []struct {
		name   string
		body   string
		setup  func()
		expect func(status int, body map[string]any)
	}{
		{
			name: "201 created",
			body: validBody,
			setup: func() {
				s.createC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusCreated, status)
				s.Equal(testCategoryID, body["id"])
			},
		},
		{
			name:  "422 invalid body",
			body:  "not-json",
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "422 invalid name",
			body: validBody,
			setup: func() {
				s.createC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CategoryResponse{}, domain.ErrInvalidCategoryName).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "404 parent not found",
			body: validBody,
			setup: func() {
				s.createC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CategoryResponse{}, domain.ErrParentNotFound).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
		{
			name: "409 duplicated name",
			body: validBody,
			setup: func() {
				s.createC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CategoryResponse{}, domain.ErrCategoryNameAlreadyExists).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusConflict, status)
			},
		},
		{
			name: "500 internal",
			body: validBody,
			setup: func() {
				s.createC.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(dtos.CategoryResponse{}, errors.New("db")).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusInternalServerError, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString(sc.body))
			req.Header.Set("Content-Type", "application/json")
			res, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			_ = json.NewDecoder(res.Body).Decode(&body)
			sc.expect(res.StatusCode, body)
		})
	}
}

func (s *CategoryHandlerSuite) TestUpdate() {
	resp := sampleResponse()
	validBody := `{"name":"Mercado","color":"blue","icon":"cart"}`

	scenarios := []struct {
		name   string
		id     string
		body   string
		setup  func()
		expect func(status int, body map[string]any)
	}{
		{
			name: "200 updated",
			id:   testCategoryID,
			body: validBody,
			setup: func() {
				s.updateC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID, mock.Anything).Return(resp, nil).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
			},
		},
		{
			name:  "422 invalid body",
			id:    testCategoryID,
			body:  "not-json",
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "422 invalid parentId / parent inactive",
			id:   testCategoryID,
			body: validBody,
			setup: func() {
				s.updateC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID, mock.Anything).Return(dtos.CategoryResponse{}, domain.ErrParentInactive).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name: "404 not found",
			id:   testCategoryID,
			body: validBody,
			setup: func() {
				s.updateC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID, mock.Anything).Return(dtos.CategoryResponse{}, domain.ErrCategoryNotFound).Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusNotFound, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("PUT", "/categories/"+sc.id, bytes.NewBufferString(sc.body))
			req.Header.Set("Content-Type", "application/json")
			res, err := s.app.Test(req)
			s.Require().NoError(err)
			var body map[string]any
			_ = json.NewDecoder(res.Body).Decode(&body)
			sc.expect(res.StatusCode, body)
		})
	}
}

func (s *CategoryHandlerSuite) TestDelete() {
	scenarios := []struct {
		name   string
		id     string
		setup  func()
		expect func(status int)
	}{
		{
			name: "204 deleted",
			id:   testCategoryID,
			setup: func() {
				s.deleteC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(nil).Once()
			},
			expect: func(status int) { s.Equal(fiber.StatusNoContent, status) },
		},
		{
			name: "404 not found",
			id:   testCategoryID,
			setup: func() {
				s.deleteC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(domain.ErrCategoryNotFound).Once()
			},
			expect: func(status int) { s.Equal(fiber.StatusNotFound, status) },
		},
		{
			name: "422 invalid id",
			id:   "not-uuid",
			setup: func() {
				s.deleteC.EXPECT().Execute(mock.Anything, mock.Anything, "not-uuid").Return(domain.ErrInvalidCategoryID).Once()
			},
			expect: func(status int) { s.Equal(fiber.StatusUnprocessableEntity, status) },
		},
		{
			name: "500 internal",
			id:   testCategoryID,
			setup: func() {
				s.deleteC.EXPECT().Execute(mock.Anything, mock.Anything, testCategoryID).Return(errors.New("db")).Once()
			},
			expect: func(status int) { s.Equal(fiber.StatusInternalServerError, status) },
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("DELETE", "/categories/"+sc.id, nil)
			res, err := s.app.Test(req)
			s.Require().NoError(err)
			sc.expect(res.StatusCode)
		})
	}
}

func (s *CategoryHandlerSuite) TestUnauthorized() {
	app := fiber.New()
	app.Get("/categories", s.sut.List)
	req := httptest.NewRequest("GET", "/categories", nil)
	res, err := app.Test(req)
	s.Require().NoError(err)
	s.Equal(fiber.StatusUnauthorized, res.StatusCode)
}
