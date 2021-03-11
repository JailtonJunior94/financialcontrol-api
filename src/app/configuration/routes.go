package configuration

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1")
	fmt.Println(v1)

	// routes.AddRankingRouter(v1)
}
