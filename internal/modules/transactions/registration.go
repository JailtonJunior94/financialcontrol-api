package transactions

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	modulehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
)

func Registration() platformmodules.ModuleRegistration {
	return platformmodules.ModuleRegistration{
		Name: "transactions",
		RegisterHTTP: func(router fiber.Router, container *bootstrapcontainer.Container) {
			modulehttp.AddTransactionRouter(router, container.TransactionController)
		},
	}
}
