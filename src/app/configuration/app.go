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

	go ioc.UpdateTransactionBill.Execute()
	go ioc.UpdateUseCase.Execute("B4351E7E-F9AC-4A84-A113-A0E159303281")
	go ioc.UpdateUseCase.Execute("FF8C5393-2C43-4AE4-92F7-42AF4DD3AF08")
	go ioc.UpdateUseCase.Execute("45DE5288-D5D0-471A-BF18-09FE1FD2FC86")
	go ioc.UpdateUseCase.Execute("4FAE4733-FB19-4F0C-A678-3C6B7588F750")

	return app
}
