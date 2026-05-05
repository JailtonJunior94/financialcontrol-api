package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
)

func MapError(err error) (int, any) {
	canonical := canonicalDomainError(err)
	switch {
	case err == nil:
		return fiber.StatusOK, nil
	case canonical == domain.ErrCardNotFound, canonical == domain.ErrFlagNotFound:
		return fiber.StatusNotFound, fiber.Map{"error": canonical.Error()}
	case canonical == domain.ErrInvalidCardID,
		canonical == domain.ErrInvalidFlagID,
		canonical == domain.ErrInvalidCardName,
		canonical == domain.ErrInvalidCardNumber,
		canonical == domain.ErrInvalidClosingDay,
		canonical == domain.ErrInvalidDueDay,
		canonical == domain.ErrInvalidBillingCycle:
		return fiber.StatusUnprocessableEntity, fiber.Map{"error": canonical.Error()}
	default:
		slog.Error("internal error in cards module", "error", err)
		return fiber.StatusInternalServerError, fiber.Map{"error": "Erro interno"}
	}
}

func canonicalDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrCardNotFound):
		return domain.ErrCardNotFound
	case errors.Is(err, domain.ErrFlagNotFound):
		return domain.ErrFlagNotFound
	case errors.Is(err, domain.ErrInvalidCardID):
		return domain.ErrInvalidCardID
	case errors.Is(err, domain.ErrInvalidFlagID):
		return domain.ErrInvalidFlagID
	case errors.Is(err, domain.ErrInvalidCardName):
		return domain.ErrInvalidCardName
	case errors.Is(err, domain.ErrInvalidCardNumber):
		return domain.ErrInvalidCardNumber
	case errors.Is(err, domain.ErrInvalidClosingDay):
		return domain.ErrInvalidClosingDay
	case errors.Is(err, domain.ErrInvalidDueDay):
		return domain.ErrInvalidDueDay
	case errors.Is(err, domain.ErrInvalidBillingCycle):
		return domain.ErrInvalidBillingCycle
	default:
		return nil
	}
}
