package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	cardsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application"

	"github.com/gofiber/fiber/v2"
)

type CardController struct {
	service        cardsapp.CardService
	claimsResolver ClaimsResolver
}

func NewCardController(service cardsapp.CardService, claimsResolver ClaimsResolver) *CardController {
	return &CardController{
		service:        service,
		claimsResolver: claimsResolver,
	}
}

func (c *CardController) Cards(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.Cards(userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *CardController) CardById(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.CardById(ctx.Params("id"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *CardController) CreateCard(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	request := new(cardsapp.CardRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	response := c.service.CreateCard(userID, request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *CardController) UpdateCard(ctx *fiber.Ctx) error {
	request := new(cardsapp.CardRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.UpdateCard(ctx.Params("id"), userID, request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *CardController) RemoveCard(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.RemoveCard(ctx.Params("id"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *CardController) inputIsValid(ctx *fiber.Ctx, request *cardsapp.CardRequest) (bool, int, any) {
	if err := ctx.BodyParser(request); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := request.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
