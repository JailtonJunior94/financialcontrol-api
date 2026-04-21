package http

import (
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"

	"github.com/gofiber/fiber/v2"
)

type CategoryController struct {
	service catalogapp.CategoryService
}

func NewCategoryController(service catalogapp.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

func (c *CategoryController) Categories(ctx *fiber.Ctx) error {
	response := c.service.Categories()
	return ctx.Status(response.StatusCode).JSON(response.Data)
}
