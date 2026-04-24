package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router, controller *TransactionController) {
	router.Get(routeconstants.Transactions, middlewares.Protected(), controller.Transactions)
	router.Get(routeconstants.TransactionDetail, middlewares.Protected(), controller.TransactionById)
	router.Post(routeconstants.Transactions, middlewares.Protected(), controller.CreateTransaction)
	router.Post(routeconstants.TransactionClone, middlewares.Protected(), controller.CloneTransaction)
	router.Get(routeconstants.TransactionIdAndItemId, middlewares.Protected(), controller.TransactionItemById)
	router.Post(routeconstants.TransactionId, middlewares.Protected(), controller.CreateTransactionItem)
	router.Put(routeconstants.TransactionIdAndItemId, middlewares.Protected(), controller.UpdateTransactionItem)
	router.Patch(routeconstants.TransactionIdAndItemId, middlewares.Protected(), controller.MarkAsPaidTransactionItem)
	router.Delete(routeconstants.TransactionIdAndItemId, middlewares.Protected(), controller.RemoveTransactionItem)
}
