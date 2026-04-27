package handlers

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authenticateUser     usecase.AuthenticateUser
	getAuthenticatedUser usecase.GetAuthenticatedUser
}

func NewAuthHandler(authenticateUser usecase.AuthenticateUser, getAuthenticatedUser usecase.GetAuthenticatedUser) *AuthHandler {
	return &AuthHandler{
		authenticateUser:     authenticateUser,
		getAuthenticatedUser: getAuthenticatedUser,
	}
}

func (h *AuthHandler) Authenticate(c *fiber.Ctx) error {
	var req dtos.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.InvalidUserOrPasswordMessage})
	}

	resp, err := h.authenticateUser.Execute(c.UserContext(), req)
	if err != nil {
		return MapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	resp, err := h.getAuthenticatedUser.Execute(c.UserContext())
	if err != nil {
		return MapError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
