package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddInvoiceRouter(router fiber.Router, controller *InvoiceController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Invoices, authmiddleware.Protected(parser), controller.Invoices)
	router.Post(pkgroutes.Invoices, authmiddleware.Protected(parser), controller.CreateInvoice)
	router.Get(pkgroutes.InvoicesById, authmiddleware.Protected(parser), controller.InvoiceById)
	router.Patch(pkgroutes.InvoicesById, authmiddleware.Protected(parser), controller.InvoiceById)
	router.Put(pkgroutes.InvoicesItems, authmiddleware.Protected(parser), controller.UpdateInvoice)
	router.Post(pkgroutes.InvoicesImport, authmiddleware.Protected(parser), controller.ImportInvoices)
	router.Delete(pkgroutes.InvoicesItems, authmiddleware.Protected(parser), controller.DeleteInvoice)
	router.Get(pkgroutes.InvoicesCategories, authmiddleware.Protected(parser), controller.InvoiceCategories)
}
