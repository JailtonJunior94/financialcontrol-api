package configuration

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func App() *fiber.App {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	environments.SetupEnvironments()

	sqlConnection := database.NewConnection()
	defer sqlConnection.Disconnect()

	SetupDependencyInjection(sqlConnection)
	SetupRoutes(app)

	return app
}
