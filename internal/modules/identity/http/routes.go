package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"

	"github.com/gofiber/fiber/v2"
)

func AddAuthRouter(router fiber.Router, controller *AuthController) {
	router.Post(routeconstants.Token, controller.Authenticate)
	router.Get("/me", controller.Me)
}

func AddUserRouter(router fiber.Router, controller *UserController) {
	router.Post(routeconstants.Users, controller.Create)
}
