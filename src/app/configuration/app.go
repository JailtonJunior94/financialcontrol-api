package configuration

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func App() *fiber.App {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	SetupEnvironments()
	SetupDependencyInjection()
	SetupRoutes(app)

	return app
}
