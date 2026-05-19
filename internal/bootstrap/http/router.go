package http

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	observabilitymiddleware "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/middleware"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules"

	"github.com/gofiber/fiber/v2"
)

type apiV1Router struct {
	container *bootstrapcontainer.Container
}

func newAPIV1Router(c *bootstrapcontainer.Container) *apiV1Router {
	return &apiV1Router{container: c}
}

// Register implements serverfiber.Router. Creates the /api/v1 group, mounts
// the FinanceMetrics middleware on the /finance prefix, then delegates route
// registration to modules.RegisterHTTP.
func (r *apiV1Router) Register(app *fiber.App) {
	v1 := app.Group("/api/v1")
	v1.Use("/finance", observabilitymiddleware.FinanceMetrics(r.container.BusinessMetrics))
	modules.RegisterHTTP(v1, r.container)
}
