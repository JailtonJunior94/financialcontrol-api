package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

// MapError maps a finance domain error to an HTTP status code and response body.
// Uses errors.Is to handle wrapped errors. Logs unexpected errors at ERROR level.
func MapError(err error) (int, any) {
	switch {
	// 400 — boundary validation sentinels (G3.a + H3.a) and VO parse errors
	case errors.Is(err, domain.ErrInvalidDateRange),
		errors.Is(err, domain.ErrInvalidPaginationLimits),
		errors.Is(err, domain.ErrInvalidISODate),
		errors.Is(err, domain.ErrSubcategoryEqualsCategory),
		errors.Is(err, domain.ErrInvalidIdempotencyKeyFormat),
		errors.Is(err, domain.ErrInvalidYearMonth),
		errors.Is(err, domain.ErrInvalidMoneyFormat),
		errors.Is(err, domain.ErrInvalidPaymentMethod),
		errors.Is(err, vos.ErrInvalidTransactionID),
		errors.Is(err, vos.ErrInvalidInvoiceID),
		errors.Is(err, vos.ErrInvalidInstallmentID),
		errors.Is(err, vos.ErrInvalidCardID),
		errors.Is(err, vos.ErrInvalidCategoryID):
		return fiber.StatusBadRequest, fiber.Map{"error": canonicalMessage(err)}

	// 404 — not found
	case errors.Is(err, domain.ErrTransactionNotFound),
		errors.Is(err, domain.ErrInvoiceNotFound),
		errors.Is(err, domain.ErrInstallmentNotFound),
		errors.Is(err, domain.ErrCardNotFound),
		errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrSubcategoryNotFound):
		return fiber.StatusNotFound, fiber.Map{"error": canonicalMessage(err)}

	// 422 — domain invariants
	case errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrDescriptionRequired),
		errors.Is(err, domain.ErrCardRequiredForPaymentMethod),
		errors.Is(err, domain.ErrCardNotActive),
		errors.Is(err, domain.ErrCategoryNotActive),
		errors.Is(err, domain.ErrSubcategoryNotActive),
		errors.Is(err, domain.ErrSubcategoryNotChildOfCategory),
		errors.Is(err, domain.ErrInstallmentCountOutOfRange),
		errors.Is(err, domain.ErrOccurredAtTooFarInFuture),
		errors.Is(err, domain.ErrInvoiceCannotPay):
		return fiber.StatusUnprocessableEntity, fiber.Map{"error": canonicalMessage(err)}

	// 409 — conflict
	case errors.Is(err, domain.ErrInvoiceAlreadyPaid),
		errors.Is(err, domain.ErrInstallmentInClosedOrPaidInvoice),
		errors.Is(err, domain.ErrRefundAlreadyExists),
		errors.Is(err, domain.ErrRefundOfRefundNotAllowed),
		errors.Is(err, domain.ErrTransactionHasDependentRefund),
		errors.Is(err, domain.ErrIdempotencyMismatch):
		return fiber.StatusConflict, fiber.Map{"error": canonicalMessage(err)}

	// 500 — unexpected errors
	default:
		slog.Error("internal error in finance module", "error", err)
		return fiber.StatusInternalServerError, fiber.Map{"error": "erro interno do servidor"}
	}
}

// canonicalMessage unwraps a sentinel and returns its canonical message.
func canonicalMessage(err error) string {
	sentinels := []error{
		domain.ErrInvalidDateRange, domain.ErrInvalidPaginationLimits, domain.ErrInvalidISODate,
		domain.ErrSubcategoryEqualsCategory, domain.ErrInvalidIdempotencyKeyFormat,
		domain.ErrInvalidYearMonth, domain.ErrInvalidMoneyFormat, domain.ErrInvalidPaymentMethod,
		vos.ErrInvalidTransactionID, vos.ErrInvalidInvoiceID, vos.ErrInvalidInstallmentID,
		vos.ErrInvalidCardID, vos.ErrInvalidCategoryID,
		domain.ErrTransactionNotFound, domain.ErrInvoiceNotFound, domain.ErrInstallmentNotFound,
		domain.ErrCardNotFound, domain.ErrCategoryNotFound, domain.ErrSubcategoryNotFound,
		domain.ErrInvalidAmount, domain.ErrDescriptionRequired, domain.ErrCardRequiredForPaymentMethod,
		domain.ErrCardNotActive, domain.ErrCategoryNotActive, domain.ErrSubcategoryNotActive,
		domain.ErrSubcategoryNotChildOfCategory, domain.ErrInstallmentCountOutOfRange,
		domain.ErrOccurredAtTooFarInFuture, domain.ErrInvoiceCannotPay,
		domain.ErrInvoiceAlreadyPaid, domain.ErrInstallmentInClosedOrPaidInvoice,
		domain.ErrRefundAlreadyExists, domain.ErrRefundOfRefundNotAllowed,
		domain.ErrTransactionHasDependentRefund, domain.ErrIdempotencyMismatch,
	}
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return s.Error()
		}
	}
	return err.Error()
}
