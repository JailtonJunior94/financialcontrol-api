package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AddFlagRouter(router fiber.Router, controller *FlagController) {
	router.Get(pkgroutes.Flags, platformsecurity.Protected(), controller.Flags)
}

func AddCategoryRouter(router fiber.Router, controller *CategoryController) {
	router.Get(pkgroutes.Categories, platformsecurity.Protected(), controller.Categories)
}
