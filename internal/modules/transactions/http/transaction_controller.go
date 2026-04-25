package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"

	"github.com/gofiber/fiber/v2"
)

type TransactionController struct {
	service        transactionsapp.TransactionAppService
	claimsResolver ClaimsResolver
}

func NewTransactionController(service transactionsapp.TransactionAppService, claimsResolver ClaimsResolver) *TransactionController {
	return &TransactionController{
		service:        service,
		claimsResolver: claimsResolver,
	}
}

func (c *TransactionController) Transactions(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.Transactions(userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) TransactionById(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.TransactionById(ctx.Params("id"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) CreateTransaction(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	request := new(transactionsapp.TransactionRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := c.service.CreateTransaction(request, userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) CloneTransaction(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.CloneTransaction(ctx.Params("transactionid"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) TransactionItemById(ctx *fiber.Ctx) error {
	response := c.service.TransactionItemById(ctx.Params("transactionid"), ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) CreateTransactionItem(ctx *fiber.Ctx) error {
	request := new(transactionsapp.TransactionItemRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.CreateTransactionItem(request, ctx.Params("transactionid"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) UpdateTransactionItem(ctx *fiber.Ctx) error {
	request := new(transactionsapp.TransactionItemRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.UpdateTransactionItem(ctx.Params("transactionid"), ctx.Params("id"), userID, request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) MarkAsPaidTransactionItem(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	var request transactionsapp.TransactionMarkAsPaid
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	response := c.service.MarkAsPaidTransactionItem(ctx.Params("transactionid"), ctx.Params("id"), userID, &request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) RemoveTransactionItem(ctx *fiber.Ctx) error {
	userID, err := c.claimsResolver.UserID(ctx.Get("Authorization"))
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.RemoveTransactionItem(ctx.Params("transactionid"), ctx.Params("id"), userID)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *TransactionController) inputIsValid(ctx *fiber.Ctx, request *transactionsapp.TransactionItemRequest) (bool, int, any) {
	if err := ctx.BodyParser(request); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := request.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
