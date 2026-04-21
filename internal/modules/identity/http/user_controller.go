package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	service identityapp.UserService
}

func NewUserController(service identityapp.UserService) *UserController {
	return &UserController{service: service}
}

func (u *UserController) Create(c *fiber.Ctx) error {
	request := new(identityapp.UserRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.service.CreateUser(request)
	return c.Status(response.StatusCode).JSON(response.Data)
}
