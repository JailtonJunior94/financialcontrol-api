package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/gofiber/fiber/v2"
)

func AddUserRouter(router fiber.Router) {
	router.Post("/users", ioc.UserController.Create)
}
