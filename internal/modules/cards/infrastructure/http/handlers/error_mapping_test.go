package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
)

func TestMapError(t *testing.T) {
	t.Parallel()
	scenarios := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   any
	}{
		{"nil error", nil, http.StatusOK, nil},
		{"ErrCardNotFound", domain.ErrCardNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrCardNotFound.Error()}},
		{"ErrFlagNotFound", domain.ErrFlagNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrFlagNotFound.Error()}},
		{"ErrCardNotFound wrapped", fmt.Errorf("ctx: %w", domain.ErrCardNotFound), http.StatusNotFound, fiber.Map{"error": domain.ErrCardNotFound.Error()}},
		{"ErrInvalidCardID", domain.ErrInvalidCardID, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCardID.Error()}},
		{"ErrInvalidFlagID", domain.ErrInvalidFlagID, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidFlagID.Error()}},
		{"ErrInvalidCardName", domain.ErrInvalidCardName, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCardName.Error()}},
		{"ErrInvalidCardNumber", domain.ErrInvalidCardNumber, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidCardNumber.Error()}},
		{"ErrInvalidClosingDay", domain.ErrInvalidClosingDay, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidClosingDay.Error()}},
		{"ErrInvalidDueDay", domain.ErrInvalidDueDay, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidDueDay.Error()}},
		{"ErrInvalidBillingCycle", domain.ErrInvalidBillingCycle, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidBillingCycle.Error()}},
		{"unknown error", errors.New("unexpected failure"), http.StatusInternalServerError, fiber.Map{"error": "Erro interno"}},
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
