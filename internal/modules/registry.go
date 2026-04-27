package modules

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"
	pkgauthmiddleware "github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func Registrations() []platformmodules.ModuleRegistration {
	return []platformmodules.ModuleRegistration{
		catalog.Registration(),
		cards.Registration(),
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
