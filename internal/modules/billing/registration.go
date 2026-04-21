package billing

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/routes"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/internal/platform/modules"

	"github.com/gofiber/fiber/v2"
)

func Registration() platformmodules.ModuleRegistration {
	return platformmodules.ModuleRegistration{
		Name: "billing",
		RegisterHTTP: func(router fiber.Router, container *bootstrapcontainer.Container) {
			routes.AddBillRouter(router, container)
		},
	}
}
