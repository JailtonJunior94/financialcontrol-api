package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	financehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http"
)

// SummaryHandler handles HTTP endpoints for the monthly financial summary.
type SummaryHandler struct {
	summary usecase.MonthlySummary
}

func NewSummaryHandler(summary usecase.MonthlySummary) *SummaryHandler {
	return &SummaryHandler{summary: summary}
}

// Get handles GET /finance/summary?year=2026&month=5.
func (h *SummaryHandler) Get(c *fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	year := parseIntQuery(c.Query("year"), 0)
	month := parseIntQuery(c.Query("month"), 0)

	period, err := vos.NewPeriod(year, month, nil)
	if err != nil {
		status, body := financehttp.MapError(domain.ErrInvalidYearMonth)
		return c.Status(status).JSON(body)
	}

	out, err := h.summary.Execute(c.UserContext(), userID, period)
	if err != nil {
		status, body := financehttp.MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}
