package handlers

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	createUser usecase.CreateUser
}

func NewUserHandler(createUser usecase.CreateUser) *UserHandler {
	return &UserHandler{createUser: createUser}
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req dtos.UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "campos obrigatórios ausentes"})
	}

	result, err := h.createUser.Execute(c.UserContext(), req)
	if err != nil {
		return MapError(c, err)
	}

	if result.Created {
		return c.Status(fiber.StatusCreated).JSON(result.User)
	}
	return c.Status(fiber.StatusOK).JSON(result.User)
}
