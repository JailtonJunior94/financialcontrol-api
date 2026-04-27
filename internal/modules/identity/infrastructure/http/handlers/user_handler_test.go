package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	ucmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase/mocks"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"

	"github.com/gofiber/fiber/v2"
)

type UserHandlerSuite struct {
	suite.Suite
	createUser *ucmocks.CreateUser
	sut        *handlers.UserHandler
	app        *fiber.App
}

func TestUserHandlerSuite(t *testing.T) { suite.Run(t, new(UserHandlerSuite)) }

func (s *UserHandlerSuite) SetupTest() {
	s.createUser = ucmocks.NewCreateUser(s.T())
	s.sut = handlers.NewUserHandler(s.createUser)

	s.app = fiber.New()
	s.app.Post("/users", s.sut.Create)
}

func (s *UserHandlerSuite) TestCreate() {
	validReq := dtos.UserRequest{Name: "Alice", Email: "alice@example.com", Password: "secret"}
	validUser := dtos.UserResponse{ID: "id1", Name: "Alice", Email: "alice@example.com", Active: true}

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
			name: "201 novo usuário criado",
			args: args{body: `{"name":"Alice","email":"alice@example.com","password":"secret"}`},
			setup: func() {
				s.createUser.EXPECT().
					Execute(mock.Anything, validReq).
					Return(usecase.CreateUserResult{User: validUser, Created: true}, nil).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusCreated, status)
				s.Equal("id1", body["id"])
				s.Equal("Alice", body["name"])
			},
		},
		{
			name: "200 idempotente — e-mail existente com mesmas credenciais",
			args: args{body: `{"name":"Alice","email":"alice@example.com","password":"secret"}`},
			setup: func() {
				s.createUser.EXPECT().
					Execute(mock.Anything, validReq).
					Return(usecase.CreateUserResult{User: validUser, Created: false}, nil).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusOK, status)
				s.Equal("id1", body["id"])
			},
		},
		{
			name: "409 e-mail existente com credenciais divergentes",
			args: args{body: `{"name":"Alice","email":"alice@example.com","password":"secret"}`},
			setup: func() {
				s.createUser.EXPECT().
					Execute(mock.Anything, validReq).
					Return(usecase.CreateUserResult{}, domain.ErrUserAlreadyExists).
					Once()
			},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusConflict, status)
				s.Contains(body, "error")
			},
		},
		{
			name:  "400 body inválido",
			args:  args{body: `not-json`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusUnprocessableEntity, status)
			},
		},
		{
			name:  "400 campos obrigatórios ausentes — name vazio",
			args:  args{body: `{"email":"alice@example.com","password":"secret"}`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name:  "400 campos obrigatórios ausentes — email vazio",
			args:  args{body: `{"name":"Alice","password":"secret"}`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
		{
			name:  "400 campos obrigatórios ausentes — password vazio",
			args:  args{body: `{"name":"Alice","email":"alice@example.com"}`},
			setup: func() {},
			expect: func(status int, body map[string]any) {
				s.Equal(fiber.StatusBadRequest, status)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(sc.args.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := s.app.Test(req)
			s.Require().NoError(err)

			var body map[string]any
			s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
			sc.expect(resp.StatusCode, body)
		})
	}
}

// compile-time guard: CreateUser mock satisfies the interface.
var _ usecase.CreateUser = (*ucmocks.CreateUser)(nil)
