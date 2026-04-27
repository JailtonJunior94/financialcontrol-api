package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router, controller *TransactionController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Transactions, authmiddleware.Protected(parser), controller.Transactions)
	router.Get(pkgroutes.TransactionDetail, authmiddleware.Protected(parser), controller.TransactionById)
	router.Post(pkgroutes.Transactions, authmiddleware.Protected(parser), controller.CreateTransaction)
	router.Post(pkgroutes.TransactionClone, authmiddleware.Protected(parser), controller.CloneTransaction)
	router.Get(pkgroutes.TransactionIdAndItemId, authmiddleware.Protected(parser), controller.TransactionItemById)
	router.Post(pkgroutes.TransactionId, authmiddleware.Protected(parser), controller.CreateTransactionItem)
	router.Put(pkgroutes.TransactionIdAndItemId, authmiddleware.Protected(parser), controller.UpdateTransactionItem)
	router.Patch(pkgroutes.TransactionIdAndItemId, authmiddleware.Protected(parser), controller.MarkAsPaidTransactionItem)
	router.Delete(pkgroutes.TransactionIdAndItemId, authmiddleware.Protected(parser), controller.RemoveTransactionItem)
}
