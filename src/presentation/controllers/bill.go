package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"

	"github.com/gofiber/fiber/v2"
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

func (u *BillController) BillItemById(c *fiber.Ctx) error {
	response := u.Service.BillItemById(c.Params("id"), c.Params("billid"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) CreateBillItem(c *fiber.Ctx) error {
	var request requests.BillItemRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	response := u.Service.CreateBillItem(&request, c.Params("billid"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) UpdateBillItem(c *fiber.Ctx) error {
	var request requests.BillItemRequest
	if err, statusCode, data := u.inputIsValid(&request, c); err {
		return c.Status(statusCode).JSON(data)
	}

	response := u.Service.UpdateBillItem(c.Params("billid"), c.Params("id"), &request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) RemoveBillItem(c *fiber.Ctx) error {
	response := u.Service.RemoveBillItem(c.Params("billid"), c.Params("id"))
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *BillController) inputIsValid(r *requests.BillItemRequest, c *fiber.Ctx) (isError bool, statusCode int, data interface{}) {
	if err := c.BodyParser(r); err != nil {
		return true, fiber.StatusUnprocessableEntity, fiber.Map{"error": customErrors.UnprocessableEntityMessage}
	}

	if err := r.IsValid(); err != nil {
		return true, fiber.StatusBadRequest, fiber.Map{"error": err.Error()}
	}

	return false, fiber.StatusOK, nil
}
