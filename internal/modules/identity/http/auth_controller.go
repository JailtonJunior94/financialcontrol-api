package http

import (
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	service        identityapp.AuthService
	claimsResolver ClaimsResolver
}

func NewAuthController(service identityapp.AuthService, claimsResolver ClaimsResolver) *AuthController {
	return &AuthController{
		service:        service,
		claimsResolver: claimsResolver,
	}
}

func (u *AuthController) Authenticate(c *fiber.Ctx) error {
	request := new(identityapp.AuthRequest)
	if err := c.BodyParser(request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": customerrors.UnprocessableEntityMessage})
	}

	if err := request.IsValid(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := u.service.Authenticate(request)
	return c.Status(response.StatusCode).JSON(response.Data)
}

func (u *AuthController) Me(c *fiber.Ctx) error {
	userID, err := u.claimsResolver.UserID(c.Get("Authorization"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": customerrors.InvalidTokenMessage})
	}

	response := u.service.Me(userID)
	return c.Status(response.StatusCode).JSON(response.Data)
}
