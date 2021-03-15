package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
)

type TransactionController struct {
	Jwt adapters.IJwtAdapter
}

func NewTransactionController(j adapters.IJwtAdapter) *TransactionController {
	return &TransactionController{Jwt: j}
}

func (u *TransactionController) Create(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	id, err := u.Jwt.ExtractClaims(authorization)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	request := new(interface{})
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	return c.Status(200).JSON(id)
}
