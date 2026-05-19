package handlers

import (
	"errors"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

const (
	msgInvalidCredentials = "usuário e/ou senha inválidos"
	msgConflict           = "recurso já cadastrado com dados divergentes"
	msgInvalidToken       = "token inválido"
	msgInternalError      = "erro interno do servidor"
)

func MapError(err error) (int, any) {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrUserNotFound):
		return fiber.StatusBadRequest, fiber.Map{"error": msgInvalidCredentials}
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return fiber.StatusConflict, fiber.Map{"error": msgConflict}
	case errors.Is(err, identitycontext.ErrNoIdentity),
		errors.Is(err, domain.ErrIdentityInvalid),
		errors.Is(err, pkgjwt.ErrInvalidToken),
		errors.Is(err, pkgjwt.ErrExpiredToken),
		errors.Is(err, pkgjwt.ErrAlgorithmMismatch):
		return fiber.StatusUnauthorized, fiber.Map{"error": msgInvalidToken}
	default:
		return fiber.StatusInternalServerError, fiber.Map{"error": msgInternalError}
	}
}
