package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddAuthRouter(router fiber.Router, controller *AuthController) {
	router.Post(pkgroutes.Token, controller.Authenticate)
	router.Get("/me", controller.Me)
}

func AddUserRouter(router fiber.Router, controller *UserController) {
	router.Post(pkgroutes.Users, controller.Create)
}
