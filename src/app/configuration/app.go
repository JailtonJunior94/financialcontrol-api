package configuration

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func App() *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://financialcontrol.netlify.app,http://localhost:3000,https://financeiro.limateixeira.site,http://financialweb-service",
	}))
	app.Use(logger.New())

	environments.SetupEnvironments()

	sqlConnection := database.NewConnection()
	ioc.SetupDependencyInjection(sqlConnection)
	SetupRoutes(app)

	return app
}
