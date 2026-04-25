package http

import (
	"fmt"
	"os"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	platformhttp "github.com/jailtonjunior94/financialcontrol-api/pkg/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewApp(container *bootstrapcontainer.Container) *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://financialcontrol.netlify.app,http://localhost:3000,https://financeiro.limateixeira.site,http://financialweb-service",
	}))
	app.Use(logger.New())

	RegisterRoutes(app, container)

	return app
}

func BuildRuntimeApp() *fiber.App {
	return NewApp(bootstrapcontainer.BuildRuntime())
}

func RunServer() error {
	app := BuildRuntimeApp()

	fmt.Printf("🚀 API is running on http://localhost:%v", os.Getenv("PORT"))
	return app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT")))
}

func RegisterRoutes(app *fiber.App, container *bootstrapcontainer.Container) {
	platformhttp.RegisterRoutes(app, container)
}
