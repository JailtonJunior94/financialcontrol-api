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

	response := u.Service.Invoices(*userId, c.Query("cardid"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *InvoiceController) CreateInvoice(c *fiber.Ctx) error {
	request := new(requests.InvoiceRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userId, err := u.Jwt.ExtractClaims(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customErrors.InvalidTokenMessage})
	}

	response := u.Service.CreateInvoice(*userId, request)
	return c.Status(response.StatusCode).JSON(response.Data)
}
