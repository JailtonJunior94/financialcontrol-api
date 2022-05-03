package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"

	"github.com/gofiber/fiber/v2"
)

type InvoiceController struct {
	Service usecases.IInvoiceService
	Jwt     adapters.IJwtAdapter
}

func NewInvoiceController(u usecases.IInvoiceService, j adapters.IJwtAdapter) *InvoiceController {
	return &InvoiceController{Service: u, Jwt: j}
}

func (u *InvoiceController) Invoices(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.Invoices(*userId, c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) InvoiceById(c *fiber.Ctx) error {
	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.InvoiceById(*userId, c.Params("cardid"), c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) InvoiceCategories(c *fiber.Ctx) error {
	var request requests.RangeDateRequest
	if err := c.QueryParser(&request); err != nil {
		c.Status(fiber.StatusUnprocessableEntity).JSON(customErrors.UnprocessableEntityMessage)
	}

	response := u.Service.InvoiceCategories(request.StartDate, request.EndDate, c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) CreateInvoice(c *fiber.Ctx) error {
	var request requests.InvoiceRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.CreateInvoice(*userId, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) UpdateInvoice(c *fiber.Ctx) error {
	var request requests.InvoiceRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.UpdateInvoice(c.Params("id"), *userId, &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) ImportInvoices(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": customErrors.UnprocessableEntityMessage})
	}

	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.ImportInvoices(*userId, file)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) inputIsValid(r *requests.InvoiceRequest, c *fiber.Ctx) (isError bool, statusCode int, data interface{}) {
	if err := c.BodyParser(r); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customErrors.UnprocessableEntityMessage}
	}

	if err := r.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
