package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/gofiber/fiber/v2"
)

func AddAuthRouter(router fiber.Router) {
	router.Post(constants.Token, ioc.AuthController.Authenticate)
	router.Get("/me", ioc.AuthController.Me)
}
