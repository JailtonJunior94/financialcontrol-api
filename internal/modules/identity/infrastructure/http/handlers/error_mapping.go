package handlers

import (
	"errors"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

const (
	msgInvalidCredentials = "Usuário e/ou senha inválidos"
	msgConflict           = "Recurso já cadastrado com dados divergentes"
	msgInvalidToken       = "Token inválido"
	msgInternalError      = "erro interno do servidor"
)

// MapError is the single translator from domain errors to HTTP responses for the identity module.
func MapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrUserNotFound):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msgInvalidCredentials})
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": msgConflict})
	case errors.Is(err, identitycontext.ErrNoIdentity),
		errors.Is(err, pkgjwt.ErrInvalidToken),
		errors.Is(err, pkgjwt.ErrExpiredToken),
		errors.Is(err, pkgjwt.ErrAlgorithmMismatch):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": msgInvalidToken})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": msgInternalError})
	}
}
