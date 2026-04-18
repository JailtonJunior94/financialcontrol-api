package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddCategoryRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Get(routeconstants.Categories, middlewares.Protected(), container.CategoryController.Categories)
}
