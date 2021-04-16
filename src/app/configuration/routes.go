package configuration

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/app/routes"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1")

	routes.AddUserRouter(v1)
	routes.AddAuthRouter(v1)
	routes.AddTransactionRouter(v1)
	routes.AddBillRouter(v1)
}
