package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"

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
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.Transactions(*userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) TransactionById(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.TransactionById(c.Params("id"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) CreateTransaction(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	request := new(requests.TransactionRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateTransaction(request, *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) CloneTransaction(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.CloneTransaction(c.Params("transactionid"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) TransactionItemById(c *fiber.Ctx) error {
	response := u.Service.TransactionItemById(c.Params("transactionid"), c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) CreateTransactionItem(c *fiber.Ctx) error {
	var request requests.TransactionItemRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.CreateTransactionItem(&request, c.Params("transactionid"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) UpdateTransactionItem(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	var request requests.TransactionItemRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	response := u.Service.UpdateTransactionItem(c.Params("transactionid"), c.Params("id"), *userID, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) MarkAsPaidTransactionItem(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	var request requests.TransactionMarkAsPaid
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	response := u.Service.MarkAsPaidTransactionItem(c.Params("transactionid"), c.Params("id"), *userID, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) RemoveTransactionItem(c *fiber.Ctx) error {
	userID, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.Service.RemoveTransactionItem(c.Params("transactionid"), c.Params("id"), *userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *TransactionController) inputIsValid(r *requests.TransactionItemRequest, c *fiber.Ctx) (isError bool, statusCode int, data interface{}) {
	if err := c.BodyParser(r); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := r.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
