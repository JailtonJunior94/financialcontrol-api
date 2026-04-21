package http

import (
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"

	"github.com/gofiber/fiber/v2"
)

type FlagController struct {
	service catalogapp.FlagService
}

func NewFlagController(service catalogapp.FlagService) *FlagController {
	return &FlagController{service: service}
}

func (c *FlagController) Flags(ctx *fiber.Ctx) error {
	response := c.service.Flags()
	return ctx.Status(response.StatusCode).JSON(response.Data)
}
