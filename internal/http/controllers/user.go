package controllers

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	Service usecases.IUserService
}

func NewUserController(u usecases.IUserService) *UserController {
	return &UserController{Service: u}
}

func (u *UserController) Create(c *fiber.Ctx) error {
	request := new(requests.UserRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.Service.CreateUser(request)
	return c.Status(response.StatusCode).JSON(response.Data)
}
