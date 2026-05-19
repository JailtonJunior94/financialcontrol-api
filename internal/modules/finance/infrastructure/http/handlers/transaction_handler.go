package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// TransactionHandler handles HTTP endpoints for finance transactions.
type TransactionHandler struct {
	create usecase.CreateTransaction
	list   usecase.ListTransactions
	get    usecase.GetTransaction
	update usecase.UpdateTransaction
	delete usecase.DeleteTransaction
	refund usecase.RefundTransaction
}

func NewTransactionHandler(
	create usecase.CreateTransaction,
	list usecase.ListTransactions,
	get usecase.GetTransaction,
	update usecase.UpdateTransaction,
	del usecase.DeleteTransaction,
	refund usecase.RefundTransaction,
) *TransactionHandler {
	return &TransactionHandler{
		create: create,
		list:   list,
		get:    get,
		update: update,
		delete: del,
		refund: refund,
	}
}

// Create handles POST /finance/transactions.
// Reads Idempotency-Key header; generates UUID server-side when absent.
// Sets X-Idempotency-Key in 201 response (decisions D2.b + E1.b).
func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	rawKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	var idempotencyKey vos.IdempotencyKey
	if rawKey == "" {
		generated := uuid.NewString()
		slog.Info("idempotency", "generated", true, "user_id", userID.String())
		idempotencyKey, err = vos.NewIdempotencyKey(generated)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro interno do servidor"})
		}
	} else {
		idempotencyKey, err = vos.NewIdempotencyKey(rawKey)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidIdempotencyKeyFormat.Error()})
		}
	}

	var req dtos.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field == "amount" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidMoneyFormat.Error()})
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := req.Validate(); err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	out, err := h.create.Execute(c.UserContext(), userID, idempotencyKey, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	c.Set("X-Idempotency-Key", idempotencyKey.Value())
	return c.Status(fiber.StatusCreated).JSON(out)
}

// List handles GET /finance/transactions.
func (h *TransactionHandler) List(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	f, err := transactionFilterFromCtx(c, userID)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	out, err := h.list.Execute(c.UserContext(), userID, f)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

// Get handles GET /finance/transactions/:id.
func (h *TransactionHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	txID, err := vos.ParseTransactionID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	out, err := h.get.Execute(c.UserContext(), userID, txID)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

// Update handles PUT /finance/transactions/:id.
func (h *TransactionHandler) Update(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	txID, err := vos.ParseTransactionID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var req dtos.UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) && typeErr.Field == "amount" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": domain.ErrInvalidMoneyFormat.Error()})
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := req.Validate(); err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}

	out, err := h.update.Execute(c.UserContext(), userID, txID, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

// Delete handles DELETE /finance/transactions/:id.
// Returns 204 No Content on success (decision I3.b).
func (h *TransactionHandler) Delete(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	txID, err := vos.ParseTransactionID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.delete.Execute(c.UserContext(), userID, txID); err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Refund handles POST /finance/transactions/:id/refund.
func (h *TransactionHandler) Refund(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	txID, err := vos.ParseTransactionID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var req dtos.RefundTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	out, err := h.refund.Execute(c.UserContext(), userID, txID, req)
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

// --- helpers ---

func userIDFromCtx(c *fiber.Ctx) (identityvo.UserID, error) {
	identity, err := identitycontext.FromContext(c.UserContext())
	if err != nil {
		return "", err
	}
	return identityvo.ParseUserID(identity.UserID)
}

func transactionFilterFromCtx(c *fiber.Ctx, userID identityvo.UserID) (filters.TransactionFilter, error) {
	pagination, err := paginationFromCtx(c)
	if err != nil {
		return filters.TransactionFilter{}, err
	}

	var txType *vos.TransactionType
	if raw := c.Query("transaction_type"); raw != "" {
		t, err := vos.ParseTransactionType(raw)
		if err != nil {
			return filters.TransactionFilter{}, err
		}
		txType = &t
	}

	var pm *vos.PaymentMethod
	if raw := c.Query("payment_method"); raw != "" {
		p, err := vos.ParsePaymentMethod(raw)
		if err != nil {
			return filters.TransactionFilter{}, err
		}
		pm = &p
	}

	var cardID *vos.CardID
	if raw := c.Query("card_id"); raw != "" {
		cid, err := vos.ParseCardID(raw)
		if err != nil {
			return filters.TransactionFilter{}, err
		}
		cardID = &cid
	}

	var categoryID *vos.CategoryID
	if raw := c.Query("category_id"); raw != "" {
		catID, err := vos.ParseCategoryID(raw)
		if err != nil {
			return filters.TransactionFilter{}, err
		}
		categoryID = &catID
	}

	invoiceStatus, err := vos.ParseInvoiceStatusFilter(c.Query("invoice_status"))
	if err != nil {
		return filters.TransactionFilter{}, err
	}

	from, err := parseOptionalDate(c.Query("from"))
	if err != nil {
		return filters.TransactionFilter{}, domain.ErrInvalidISODate
	}

	to, err := parseOptionalDate(c.Query("to"))
	if err != nil {
		return filters.TransactionFilter{}, domain.ErrInvalidISODate
	}

	return filters.NewTransactionFilter(
		userID, txType, pm, cardID, categoryID,
		invoiceStatus, c.Query("description_contains"),
		from, to, pagination,
	)
}

func paginationFromCtx(c *fiber.Ctx) (vos.Pagination, error) {
	page := parseIntQuery(c.Query("page"), 0)
	pageSize := parseIntQuery(c.Query("page_size"), 0)
	return vos.NewPagination(page, pageSize)
}

func parseIntQuery(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func parseOptionalDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
