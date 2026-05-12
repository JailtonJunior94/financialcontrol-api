package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	financehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http"
)

// InstallmentHandler handles HTTP endpoints for finance installments.
type InstallmentHandler struct {
	anticipate usecase.AnticipateInstallment
}

func NewInstallmentHandler(anticipate usecase.AnticipateInstallment) *InstallmentHandler {
	return &InstallmentHandler{anticipate: anticipate}
}

// Anticipate handles POST /finance/installments/:id/anticipate.
// The installment ID is taken from the path param :id.
// The transaction_id is provided in the JSON body.
func (h *InstallmentHandler) Anticipate(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	installmentID, err := vos.ParseInstallmentID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var body struct {
		TransactionID string `json:"transaction_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.TransactionID == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "transaction_id é obrigatório"})
	}

	txID, err := vos.ParseTransactionID(body.TransactionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	out, err := h.anticipate.Execute(c.UserContext(), userID, txID, installmentID)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}
