package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router, controller *InvoiceController) {
	router.Get(pkgroutes.Invoices, platformsecurity.Protected(), controller.Invoices)
	router.Post(pkgroutes.Invoices, platformsecurity.Protected(), controller.CreateInvoice)
	router.Get(pkgroutes.InvoicesById, platformsecurity.Protected(), controller.InvoiceById)
	router.Patch(pkgroutes.InvoicesById, platformsecurity.Protected(), controller.InvoiceById)
	router.Put(pkgroutes.InvoicesItems, platformsecurity.Protected(), controller.UpdateInvoice)
	router.Post(pkgroutes.InvoicesImport, platformsecurity.Protected(), controller.ImportInvoices)
	router.Delete(pkgroutes.InvoicesItems, platformsecurity.Protected(), controller.DeleteInvoice)
	router.Get(pkgroutes.InvoicesCategories, platformsecurity.Protected(), controller.InvoiceCategories)
}
