package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/gofiber/fiber/v2"
)

func AddAuthRouter(router fiber.Router) {
	router.Post("/token", ioc.AuthController.Authenticate)
}
