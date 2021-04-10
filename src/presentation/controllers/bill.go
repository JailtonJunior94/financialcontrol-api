package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type BillController struct {
	Service usecases.IBillService
}

func NewBillController(u usecases.IBillService) *BillController {
	return &BillController{Service: u}
}

func (u *BillController) Bills(c *fiber.Ctx) error {
	response := u.Service.Bills()
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) BillById(c *fiber.Ctx) error {
	response := u.Service.BillById(c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) CreateBill(c *fiber.Ctx) error {
	request := new(requests.BillRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateBill(request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) CreateBillItem(c *fiber.Ctx) error {
	request := new(requests.BillItemRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customErrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateBillItem(request, c.Params("billid"))
	return c.Status(response.StatusCode).JSON(response.Data)
}
