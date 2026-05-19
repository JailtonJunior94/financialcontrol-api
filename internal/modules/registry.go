package modules

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	pkgauthmiddleware "github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
)

func Registrations() []platformmodules.ModuleRegistration {
	return []platformmodules.ModuleRegistration{}
}

func RegisterHTTP(router fiber.Router, container *bootstrapcontainer.Container) {
	protected := pkgauthmiddleware.Protected(container.JwtParser)

	container.IdentityModule.RegisterHTTP(router, protected)
	container.CardsModule.RegisterHTTP(router, protected)
	container.CategoriesModule.RegisterHTTP(router, protected)
	container.FinanceModule.RegisterHTTP(router, protected)
}
