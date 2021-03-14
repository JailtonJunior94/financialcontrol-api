package configuration

import (
	"fmt"

	"github.com/jailtonjunior94/financialcontrol-api/src/app/routes"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1")
	fmt.Println(v1)

	routes.AddUserRouter(v1)
}
