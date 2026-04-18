package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"

	"github.com/gofiber/fiber/v2"
)

func AddAuthRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Post(routeconstants.Token, container.AuthController.Authenticate)
	router.Get("/me", container.AuthController.Me)
}
