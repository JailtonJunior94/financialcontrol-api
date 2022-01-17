package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router) {
	router.Get(constants.Invoices, middlewares.Protected(), ioc.InvoiceController.Invoices)
	router.Get(constants.InvoicesById, middlewares.Protected(), ioc.InvoiceController.InvoiceById)
	router.Get(constants.InvoicesCategories, middlewares.Protected(), ioc.InvoiceController.InvoiceCategories)
	router.Post(constants.Invoices, middlewares.Protected(), ioc.InvoiceController.CreateInvoice)

	router.Post(constants.InvoicesImport, middlewares.Protected(), ioc.InvoiceController.ImportInvoices)
}
