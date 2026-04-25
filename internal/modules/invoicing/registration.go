package invoicing

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	modulehttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"

	"github.com/gofiber/fiber/v2"
)

func Registration() platformmodules.ModuleRegistration {
	return platformmodules.ModuleRegistration{
		Name: "invoicing",
		RegisterHTTP: func(router fiber.Router, container *bootstrapcontainer.Container) {
			modulehttp.AddInvoiceRouter(router, container.InvoiceController)
		},
	}
}
