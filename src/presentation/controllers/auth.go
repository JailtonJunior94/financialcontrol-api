package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	Service usecases.IAuthService
}

func NewAuthController(u usecases.IAuthService) *AuthController {
	return &AuthController{Service: u}
}

func (u *AuthController) Authenticate(c *fiber.Ctx) error {
	request := new(requests.AuthRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.Authenticate(request)
	return c.Status(response.StatusCode).JSON(response.Data)
}
