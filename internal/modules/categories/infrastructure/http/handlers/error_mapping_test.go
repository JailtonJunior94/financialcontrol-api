package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
)

func TestMapError(t *testing.T) {
	t.Parallel()
	scenarios := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   any
	}{
		{"nil", nil, http.StatusOK, nil},
		{"ErrCategoryNotFound", domain.ErrCategoryNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrCategoryNotFound.Error()}},
		{"ErrParentNotFound", domain.ErrParentNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrParentNotFound.Error()}},
		{"ErrCategoryNotFound wrapped", fmt.Errorf("ctx: %w", domain.ErrCategoryNotFound), http.StatusNotFound, fiber.Map{"error": domain.ErrCategoryNotFound.Error()}},
		{"ErrCategoryNameAlreadyExists", domain.ErrCategoryNameAlreadyExists, http.StatusConflict, fiber.Map{"error": domain.ErrCategoryNameAlreadyExists.Error()}},
		{"ErrInvalidCategoryID", domain.ErrInvalidCategoryID, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCategoryID.Error()}},
		{"ErrInvalidCategoryName", domain.ErrInvalidCategoryName, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCategoryName.Error()}},
		{"ErrInvalidCategoryColor", domain.ErrInvalidCategoryColor, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCategoryColor.Error()}},
		{"ErrInvalidCategoryIcon", domain.ErrInvalidCategoryIcon, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCategoryIcon.Error()}},
		{"ErrSubcategoryDepthExceeded", domain.ErrSubcategoryDepthExceeded, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrSubcategoryDepthExceeded.Error()}},
		{"ErrParentInactive", domain.ErrParentInactive, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrParentInactive.Error()}},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, fiber.Map{"error": "Erro interno"}},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			status, body := handlers.MapError(sc.err)
			assert.Equal(t, sc.expectedStatus, status)
			assert.Equal(t, sc.expectedBody, body)
		})
	}
}
