package handlers

import (
	"log/slog"

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

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.InvalidUserOrPasswordMessage})
	}

	resp, err := h.authenticateUser.Execute(c.UserContext(), req)
	if err != nil {
		warnAuthFailure(c, err)
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	resp, err := h.getAuthenticatedUser.Execute(c.UserContext())
	if err != nil {
		warnAuthFailure(c, err)
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// warnAuthFailure logs an auth-related failure with non-PII fields only (RNF-08).
// Records route, request ID and error reason; never email, token, claims or password.
func warnAuthFailure(c *fiber.Ctx, err error) {
	slog.Warn("identity auth failure",
		"route", c.Path(),
		"reason", err.Error(),
		"request_id", c.Get("X-Request-ID"),
	)
}
