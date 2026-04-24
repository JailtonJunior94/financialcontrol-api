package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"

	"github.com/gofiber/fiber/v2"
)

type BillController struct {
	service billingapp.BillService
}

func NewBillController(service billingapp.BillService) *BillController {
	return &BillController{service: service}
}

func (c *BillController) Bills(ctx *fiber.Ctx) error {
	response := c.service.Bills()
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) BillById(ctx *fiber.Ctx) error {
	response := c.service.BillById(ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) CreateBill(ctx *fiber.Ctx) error {
	request := new(billingapp.BillRequest)
	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := c.service.CreateBill(request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) BillItemById(ctx *fiber.Ctx) error {
	response := c.service.BillItemById(ctx.Params("id"), ctx.Params("billid"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) CreateBillItem(ctx *fiber.Ctx) error {
	request := new(billingapp.BillItemRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	response := c.service.CreateBillItem(request, ctx.Params("billid"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) UpdateBillItem(ctx *fiber.Ctx) error {
	request := new(billingapp.BillItemRequest)
	if isError, statusCode, data := c.inputIsValid(ctx, request); isError {
		return ctx.Status(statusCode).JSON(data)
	}

	response := c.service.UpdateBillItem(ctx.Params("billid"), ctx.Params("id"), request)
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) RemoveBillItem(ctx *fiber.Ctx) error {
	response := c.service.RemoveBillItem(ctx.Params("billid"), ctx.Params("id"))
	return ctx.Status(response.StatusCode).JSON(response.Data)
}

func (c *BillController) inputIsValid(ctx *fiber.Ctx, request *billingapp.BillItemRequest) (bool, int, any) {
	if err := ctx.BodyParser(request); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customerrors.UnprocessableEntityMessage}
	}

	if err := request.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
