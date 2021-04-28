package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

type CategoryController struct {
	Service usecases.ICategoryService
}

func NewCategoryController(u usecases.ICategoryService) *CategoryController {
	return &CategoryController{Service: u}
}

func (u *CategoryController) Categories(c *fiber.Ctx) error {
	response := u.Service.Categories()
	return c.Status(response.StatusCode).JSON(response.Data)
}
