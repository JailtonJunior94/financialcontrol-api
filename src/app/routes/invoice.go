package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router) {
	router.Get(constants.Invoices, middlewares.Protected(), ioc.InvoiceController.Invoices)
	router.Post(constants.Invoices, middlewares.Protected(), ioc.InvoiceController.CreateInvoice)
	router.Get(constants.InvoicesById, middlewares.Protected(), ioc.InvoiceController.InvoiceById)
	router.Patch(constants.InvoicesById, middlewares.Protected(), ioc.InvoiceController.InvoiceById)
	router.Put(constants.InvoicesItems, middlewares.Protected(), ioc.InvoiceController.UpdateInvoice)
	router.Post(constants.InvoicesImport, middlewares.Protected(), ioc.InvoiceController.ImportInvoices)
	router.Delete(constants.InvoicesItems, middlewares.Protected(), ioc.InvoiceController.DeleteInvoice)
	router.Get(constants.InvoicesCategories, middlewares.Protected(), ioc.InvoiceController.InvoiceCategories)
}
