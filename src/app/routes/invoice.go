package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router) {
	router.Post(constants.Invoices, middlewares.Protected(), ioc.InvoiceController.CreateInvoice)
}
