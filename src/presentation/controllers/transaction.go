package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"

	"github.com/gofiber/fiber/v2"
)

type TransactionController struct {
	Jwt     adapters.IJwtAdapter
	Service usecases.ITransactionService
}

func NewTransactionController(j adapters.IJwtAdapter, s usecases.ITransactionService) *TransactionController {
	return &TransactionController{Jwt: j, Service: s}
}

func (u *TransactionController) Transactions(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.Transactions(*userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) TransactionById(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.TransactionById(c.Params("id"), *userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) CreateTransaction(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	request := new(requests.TransactionRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateTransaction(request, *userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) TransactionItemById(c *fiber.Ctx) error {
	response := u.Service.TransactionItemById(c.Params("transactionid"), c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) CreateTransactionItem(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	request := new(requests.TransactionItemRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateTransactionItem(request, c.Params("transactionid"), *userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) UpdateTransactionItem(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	request := new(requests.TransactionItemRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.UpdateTransactionItem(c.Params("transactionid"), c.Params("id"), *userId, request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) RemoveTransactionItem(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.RemoveTransactionItem(c.Params("transactionid"), c.Params("id"), *userId)
	return c.Status(response.StatusCode).JSON(response.Data)
}
