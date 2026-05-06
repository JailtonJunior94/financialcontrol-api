package modules

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions"
	pkgauthmiddleware "github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func Registrations() []platformmodules.ModuleRegistration {
	return []platformmodules.ModuleRegistration{
		billing.Registration(),
		transactions.Registration(),
		invoicing.Registration(),
		planning.Registration(),
	}
}

func RegisterHTTP(router fiber.Router, container *bootstrapcontainer.Container) {
	container.IdentityModule.RegisterHTTP(
		router,
		pkgauthmiddleware.Protected(container.JwtParser),
	)

	container.CardsModule.RegisterHTTP(router, container.JwtParser)
	container.CategoriesModule.RegisterHTTP(router, container.JwtParser)

	for _, registration := range Registrations() {
		if registration.RegisterHTTP == nil {
			continue
		}

		registration.RegisterHTTP(router, container)
	}
}

func RegisterCLI(root *cobra.Command, runners platformmodules.CLIRunner) {
	for _, registration := range Registrations() {
		if registration.RegisterCLI == nil {
			continue
		}

		registration.RegisterCLI(root, runners)
	}
}
