package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"

	"github.com/gofiber/fiber/v2"
)

func Register(router fiber.Router, container *bootstrapcontainer.Container) {
	AddUserRouter(router, container)
	AddAuthRouter(router, container)
	AddTransactionRouter(router, container)
	AddBillRouter(router, container)
	AddFlagRouter(router, container)
	AddCardRouter(router, container)
	AddInvoiceRouter(router, container)
	AddCategoryRouter(router, container)
}
