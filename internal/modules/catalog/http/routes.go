package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddFlagRouter(router fiber.Router, controller *FlagController) {
	router.Get(routeconstants.Flags, middlewares.Protected(), controller.Flags)
}

func AddCategoryRouter(router fiber.Router, controller *CategoryController) {
	router.Get(routeconstants.Categories, middlewares.Protected(), controller.Categories)
}
