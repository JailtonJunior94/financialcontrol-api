package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router, controller *TransactionController) {
	router.Get(pkgroutes.Transactions, platformsecurity.Protected(), controller.Transactions)
	router.Get(pkgroutes.TransactionDetail, platformsecurity.Protected(), controller.TransactionById)
	router.Post(pkgroutes.Transactions, platformsecurity.Protected(), controller.CreateTransaction)
	router.Post(pkgroutes.TransactionClone, platformsecurity.Protected(), controller.CloneTransaction)
	router.Get(pkgroutes.TransactionIdAndItemId, platformsecurity.Protected(), controller.TransactionItemById)
	router.Post(pkgroutes.TransactionId, platformsecurity.Protected(), controller.CreateTransactionItem)
	router.Put(pkgroutes.TransactionIdAndItemId, platformsecurity.Protected(), controller.UpdateTransactionItem)
	router.Patch(pkgroutes.TransactionIdAndItemId, platformsecurity.Protected(), controller.MarkAsPaidTransactionItem)
	router.Delete(pkgroutes.TransactionIdAndItemId, platformsecurity.Protected(), controller.RemoveTransactionItem)
}
