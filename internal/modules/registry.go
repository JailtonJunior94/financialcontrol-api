package modules

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	pkgauthmiddleware "github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func Registrations() []platformmodules.ModuleRegistration {
	return []platformmodules.ModuleRegistration{}
}

func RegisterHTTP(router fiber.Router, container *bootstrapcontainer.Container) {
	container.IdentityModule.RegisterHTTP(
		router,
		pkgauthmiddleware.Protected(container.JwtParser),
	)

	container.CardsModule.RegisterHTTP(router, container.JwtParser)
	container.CategoriesModule.RegisterHTTP(router, container.JwtParser)
	container.FinanceModule.RegisterHTTP(router, container.JwtParser)
}

// RegisterCLI is kept for interface compatibility. All legacy CLI hooks were removed
// together with billing, transactions, invoicing, and planning/sync in task 9.0.
func RegisterCLI(_ *cobra.Command, _ platformmodules.CLIRunner) {}
