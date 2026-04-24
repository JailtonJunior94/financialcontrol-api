package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router, controller *InvoiceController) {
	router.Get(routeconstants.Invoices, middlewares.Protected(), controller.Invoices)
	router.Post(routeconstants.Invoices, middlewares.Protected(), controller.CreateInvoice)
	router.Get(routeconstants.InvoicesById, middlewares.Protected(), controller.InvoiceById)
	router.Patch(routeconstants.InvoicesById, middlewares.Protected(), controller.InvoiceById)
	router.Put(routeconstants.InvoicesItems, middlewares.Protected(), controller.UpdateInvoice)
	router.Post(routeconstants.InvoicesImport, middlewares.Protected(), controller.ImportInvoices)
	router.Delete(routeconstants.InvoicesItems, middlewares.Protected(), controller.DeleteInvoice)
	router.Get(routeconstants.InvoicesCategories, middlewares.Protected(), controller.InvoiceCategories)
}
