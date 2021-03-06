package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/gofiber/fiber/v2"
)

func AddUserRouter(router fiber.Router) {
	router.Post(constants.Users, ioc.UserController.Create)
}
