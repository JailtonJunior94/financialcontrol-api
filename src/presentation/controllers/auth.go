package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	Service usecases.IAuthService
	Jwt     adapters.IJwtAdapter
}

func NewAuthController(u usecases.IAuthService, j adapters.IJwtAdapter) *AuthController {
	return &AuthController{Service: u, Jwt: j}
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

func (u *AuthController) Me(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.Me(*userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}
