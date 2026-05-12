package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
)

func MapError(err error) (int, any) {
	canonical := canonicalDomainError(err)
	switch {
	case err == nil:
		return fiber.StatusOK, nil
	case canonical == domain.ErrCategoryNotFound, canonical == domain.ErrParentNotFound:
		return fiber.StatusNotFound, fiber.Map{"error": canonical.Error()}
	case canonical == domain.ErrCategoryNameAlreadyExists:
		return fiber.StatusConflict, fiber.Map{"error": canonical.Error()}
	case canonical == domain.ErrInvalidCategoryID,
		canonical == domain.ErrInvalidCategoryName,
		canonical == domain.ErrCategoryHierarchyUnsupported,
		canonical == domain.ErrInvalidCategoryColor,
		canonical == domain.ErrInvalidCategoryIcon,
		canonical == domain.ErrSubcategoryDepthExceeded,
		canonical == domain.ErrParentInactive:
		return fiber.StatusUnprocessableEntity, fiber.Map{"error": canonical.Error()}
	default:
		slog.Error("internal error in categories module", "module", "categories", "error", err)
		return fiber.StatusInternalServerError, fiber.Map{"error": "Erro interno"}
	}
}

func canonicalDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrCategoryNotFound):
		return domain.ErrCategoryNotFound
	case errors.Is(err, domain.ErrParentNotFound):
		return domain.ErrParentNotFound
	case errors.Is(err, domain.ErrCategoryNameAlreadyExists):
		return domain.ErrCategoryNameAlreadyExists
	case errors.Is(err, domain.ErrInvalidCategoryID):
		return domain.ErrInvalidCategoryID
	case errors.Is(err, domain.ErrInvalidCategoryName):
		return domain.ErrInvalidCategoryName
	case errors.Is(err, domain.ErrCategoryHierarchyUnsupported):
		return domain.ErrCategoryHierarchyUnsupported
	case errors.Is(err, domain.ErrInvalidCategoryColor):
		return domain.ErrInvalidCategoryColor
	case errors.Is(err, domain.ErrInvalidCategoryIcon):
		return domain.ErrInvalidCategoryIcon
	case errors.Is(err, domain.ErrSubcategoryDepthExceeded):
		return domain.ErrSubcategoryDepthExceeded
	case errors.Is(err, domain.ErrParentInactive):
		return domain.ErrParentInactive
	default:
		return nil
	}
}
