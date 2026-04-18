package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Get(routeconstants.Invoices, middlewares.Protected(), container.InvoiceController.Invoices)
	router.Post(routeconstants.Invoices, middlewares.Protected(), container.InvoiceController.CreateInvoice)
	router.Get(routeconstants.InvoicesById, middlewares.Protected(), container.InvoiceController.InvoiceById)
	router.Patch(routeconstants.InvoicesById, middlewares.Protected(), container.InvoiceController.InvoiceById)
	router.Put(routeconstants.InvoicesItems, middlewares.Protected(), container.InvoiceController.UpdateInvoice)
	router.Post(routeconstants.InvoicesImport, middlewares.Protected(), container.InvoiceController.ImportInvoices)
	router.Delete(routeconstants.InvoicesItems, middlewares.Protected(), container.InvoiceController.DeleteInvoice)
	router.Get(routeconstants.InvoicesCategories, middlewares.Protected(), container.InvoiceController.InvoiceCategories)
}
