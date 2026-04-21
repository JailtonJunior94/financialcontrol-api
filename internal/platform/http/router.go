package http

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, container *bootstrapcontainer.Container) {
	v1 := app.Group("/api/v1")
	modules.RegisterHTTP(v1, container)
}
