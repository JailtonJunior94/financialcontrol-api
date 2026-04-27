package catalog

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	modulehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
)

func Registration() platformmodules.ModuleRegistration {
	return platformmodules.ModuleRegistration{
		Name: "catalog",
		RegisterHTTP: func(router fiber.Router, container *bootstrapcontainer.Container) {
			modulehttp.AddFlagRouter(router, container.FlagController, container.JwtParser)
			modulehttp.AddCategoryRouter(router, container.CategoryController, container.JwtParser)
		},
	}
}
