package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"

	"github.com/gofiber/fiber/v2"
)

func AddUserRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Post(routeconstants.Users, container.UserController.Create)
}
