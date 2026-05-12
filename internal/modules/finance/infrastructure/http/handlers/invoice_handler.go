package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	financehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// InvoiceHandler handles HTTP endpoints for finance invoices.
type InvoiceHandler struct {
	list usecase.ListInvoices
	get  usecase.GetInvoice
	pay  usecase.PayInvoice
}

func NewInvoiceHandler(
	list usecase.ListInvoices,
	get usecase.GetInvoice,
	pay usecase.PayInvoice,
) *InvoiceHandler {
	return &InvoiceHandler{list: list, get: get, pay: pay}
}

// List handles GET /finance/invoices.
func (h *InvoiceHandler) List(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	f, err := buildInvoiceFilter(c, userID)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}

	out, err := h.list.Execute(c.UserContext(), userID, f)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

// Get handles GET /finance/invoices/:id.
func (h *InvoiceHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	invoiceID, err := vos.ParseInvoiceID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	out, err := h.get.Execute(c.UserContext(), userID, invoiceID)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

// Pay handles PATCH /finance/invoices/:id/pay.
func (h *InvoiceHandler) Pay(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	invoiceID, err := vos.ParseInvoiceID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var req dtos.PayInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "corpo da requisição inválido"})
	}

	if err := req.Validate(); err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}

	out, err := h.pay.Execute(c.UserContext(), userID, invoiceID, req)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func buildInvoiceFilter(c *fiber.Ctx, userID identityvo.UserID) (filters.InvoiceFilter, error) {
	pagination, err := paginationFromCtx(c)
	if err != nil {
		return filters.InvoiceFilter{}, err
	}

	var cardID *vos.CardID
	if raw := c.Query("card_id"); raw != "" {
		cid, parseErr := vos.ParseCardID(raw)
		if parseErr != nil {
			return filters.InvoiceFilter{}, parseErr
		}
		cardID = &cid
	}

	var state *vos.InvoiceState
	if raw := c.Query("state"); raw != "" {
		s, parseErr := vos.ParseInvoiceState(raw)
		if parseErr != nil {
			return filters.InvoiceFilter{}, domain.ErrInvalidISODate
		}
		state = &s
	}

	from, err := parseOptionalDate(c.Query("from"))
	if err != nil {
		return filters.InvoiceFilter{}, domain.ErrInvalidISODate
	}

	to, err := parseOptionalDate(c.Query("to"))
	if err != nil {
		return filters.InvoiceFilter{}, domain.ErrInvalidISODate
	}

	return filters.NewInvoiceFilter(userID, cardID, state, from, to, pagination)
}
