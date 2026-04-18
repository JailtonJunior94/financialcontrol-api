package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"

	"github.com/gofiber/fiber/v2"
)

type CardController struct {
	Jwt     adapters.IJwtAdapter
	Service usecases.ICardService
}

func NewCardController(u usecases.ICardService, j adapters.IJwtAdapter) *CardController {
	return &CardController{Service: u, Jwt: j}
}

func (u *CardController) Cards(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.Cards(*userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *CardController) CardById(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.CardById(c.Params("id"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *CardController) CreateCard(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	var request requests.CardRequest
	if err, statusCode, data := u.inputIsValid(c, &request); err {
		return c.Status(statusCode).JSON(data)
	}

	response := u.Service.CreateCard(*userID, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *CardController) UpdateCard(c *fiber.Ctx) error {
	var request requests.CardRequest
	if err, statusCode, data := u.inputIsValid(c, &request); err {
		return c.Status(statusCode).JSON(data)
	}

	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.UpdateCard(c.Params("id"), *userID, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *CardController) RemoveCard(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.RemoveCard(c.Params("id"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *CardController) inputIsValid(c *fiber.Ctx, r *requests.CardRequest) (isError bool, statusCode int, data interface{}) {
	if err := c.BodyParser(r); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := r.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
