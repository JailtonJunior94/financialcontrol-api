package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
)

type FlagHandler struct {
	listFlags usecase.ListFlags
}

func NewFlagHandler(listFlags usecase.ListFlags) *FlagHandler {
	return &FlagHandler{listFlags: listFlags}
}

func (h *FlagHandler) List(c *fiber.Ctx) error {
	_, err := userIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	out, err := h.listFlags.Execute(c.UserContext())
	if err != nil {
		status, body := MapError(err)
		return c.Status(status).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}
