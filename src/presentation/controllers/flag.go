package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

type FlagController struct {
	Service usecases.IFlagService
}

func NewFlagController(u usecases.IFlagService) *FlagController {
	return &FlagController{Service: u}
}

func (u *FlagController) Flags(c *fiber.Ctx) error {
	response := u.Service.Flags()
	return c.Status(response.StatusCode).JSON(response.Data)
}
