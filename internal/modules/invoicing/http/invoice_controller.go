package http

import (
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"

	"github.com/gofiber/fiber/v2"
)

type InvoiceController struct {
	service invoicingapp.InvoiceService
}

func NewInvoiceController(service invoicingapp.InvoiceService) *InvoiceController {
	return &InvoiceController{service: service}
}

func (c *InvoiceController) Invoices(ctx *fiber.Ctx) error {
	id, err := identitycontext.FromContext(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.Invoices(id.UserID, ctx.Query("cardId"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) InvoiceById(ctx *fiber.Ctx) error {
	id, err := identitycontext.FromContext(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.InvoiceById(id.UserID, ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) InvoiceCategories(ctx *fiber.Ctx) error {
	var request invoicingapp.RangeDateRequest
	if err := ctx.QueryParser(&request); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(customerrors.UnprocessableEntityMessage)
	}

	response := c.service.InvoiceCategories(request.StartDate, request.EndDate, ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) CreateInvoice(ctx *fiber.Ctx) error {
	request := new(invoicingapp.InvoiceRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	id, err := identitycontext.FromContext(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.CreateInvoice(id.UserID, request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) UpdateInvoice(ctx *fiber.Ctx) error {
	request := new(invoicingapp.InvoiceRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	id, err := identitycontext.FromContext(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.UpdateInvoice(ctx.Params("id"), id.UserID, request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) DeleteInvoice(ctx *fiber.Ctx) error {
	response := c.service.DeleteInvoiceItem(ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) ImportInvoices(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": customerrors.UnprocessableEntityMessage})
	}

	id, err := identitycontext.FromContext(ctx.UserContext())
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := c.service.ImportInvoices(id.UserID, file)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *InvoiceController) inputIsValid(ctx *fiber.Ctx, request *invoicingapp.InvoiceRequest) (bool, int, any) {
	if err := ctx.BodyParser(request); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := request.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
