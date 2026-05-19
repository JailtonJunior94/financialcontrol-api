package handlers_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   any
	}{
		// 400 — boundary validation
		{"ErrInvalidDateRange", domain.ErrInvalidDateRange, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidDateRange.Error()}},
		{"ErrInvalidPaginationLimits", domain.ErrInvalidPaginationLimits, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidPaginationLimits.Error()}},
		{"ErrInvalidISODate", domain.ErrInvalidISODate, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidISODate.Error()}},
		{"ErrSubcategoryEqualsCategory", domain.ErrSubcategoryEqualsCategory, http.StatusBadRequest, fiber.Map{"error": domain.ErrSubcategoryEqualsCategory.Error()}},
		{"ErrInvalidIdempotencyKeyFormat", domain.ErrInvalidIdempotencyKeyFormat, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidIdempotencyKeyFormat.Error()}},
		{"ErrInvalidYearMonth", domain.ErrInvalidYearMonth, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidYearMonth.Error()}},
		{"ErrInvalidMoneyFormat", domain.ErrInvalidMoneyFormat, http.StatusBadRequest, fiber.Map{"error": domain.ErrInvalidMoneyFormat.Error()}},

		// 404 — not found
		{"ErrTransactionNotFound", domain.ErrTransactionNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrTransactionNotFound.Error()}},
		{"ErrInvoiceNotFound", domain.ErrInvoiceNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrInvoiceNotFound.Error()}},
		{"ErrInstallmentNotFound", domain.ErrInstallmentNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrInstallmentNotFound.Error()}},
		{"ErrCardNotFound", domain.ErrCardNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrCardNotFound.Error()}},
		{"ErrCategoryNotFound", domain.ErrCategoryNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrCategoryNotFound.Error()}},
		{"ErrSubcategoryNotFound", domain.ErrSubcategoryNotFound, http.StatusNotFound, fiber.Map{"error": domain.ErrSubcategoryNotFound.Error()}},
		// wrapped 404
		{"ErrTransactionNotFound wrapped", fmt.Errorf("usecase: %w", domain.ErrTransactionNotFound), http.StatusNotFound, fiber.Map{"error": domain.ErrTransactionNotFound.Error()}},

		// 422 — domain invariants
		{"ErrInvalidAmount", domain.ErrInvalidAmount, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvalidAmount.Error()}},
		{"ErrDescriptionRequired", domain.ErrDescriptionRequired, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrDescriptionRequired.Error()}},
		{"ErrCardRequiredForPaymentMethod", domain.ErrCardRequiredForPaymentMethod, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrCardRequiredForPaymentMethod.Error()}},
		{"ErrCardNotActive", domain.ErrCardNotActive, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrCardNotActive.Error()}},
		{"ErrCategoryNotActive", domain.ErrCategoryNotActive, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrCategoryNotActive.Error()}},
		{"ErrSubcategoryNotActive", domain.ErrSubcategoryNotActive, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrSubcategoryNotActive.Error()}},
		{"ErrSubcategoryNotChildOfCategory", domain.ErrSubcategoryNotChildOfCategory, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrSubcategoryNotChildOfCategory.Error()}},
		{"ErrInstallmentCountOutOfRange", domain.ErrInstallmentCountOutOfRange, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInstallmentCountOutOfRange.Error()}},
		{"ErrOccurredAtTooFarInFuture", domain.ErrOccurredAtTooFarInFuture, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrOccurredAtTooFarInFuture.Error()}},
		{"ErrInvoiceCannotPay", domain.ErrInvoiceCannotPay, http.StatusUnprocessableEntity, fiber.Map{"error": domain.ErrInvoiceCannotPay.Error()}},

		// 409 — conflict
		{"ErrInvoiceAlreadyPaid", domain.ErrInvoiceAlreadyPaid, http.StatusConflict, fiber.Map{"error": domain.ErrInvoiceAlreadyPaid.Error()}},
		{"ErrInstallmentInClosedOrPaidInvoice", domain.ErrInstallmentInClosedOrPaidInvoice, http.StatusConflict, fiber.Map{"error": domain.ErrInstallmentInClosedOrPaidInvoice.Error()}},
		{"ErrRefundAlreadyExists", domain.ErrRefundAlreadyExists, http.StatusConflict, fiber.Map{"error": domain.ErrRefundAlreadyExists.Error()}},
		{"ErrRefundOfRefundNotAllowed", domain.ErrRefundOfRefundNotAllowed, http.StatusConflict, fiber.Map{"error": domain.ErrRefundOfRefundNotAllowed.Error()}},
		{"ErrTransactionHasDependentRefund", domain.ErrTransactionHasDependentRefund, http.StatusConflict, fiber.Map{"error": domain.ErrTransactionHasDependentRefund.Error()}},
		{"ErrIdempotencyMismatch", domain.ErrIdempotencyMismatch, http.StatusConflict, fiber.Map{"error": domain.ErrIdempotencyMismatch.Error()}},

		// 500 — unknown
		{"unknown error", errors.New("unexpected db failure"), http.StatusInternalServerError, fiber.Map{"error": "erro interno do servidor"}},
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
