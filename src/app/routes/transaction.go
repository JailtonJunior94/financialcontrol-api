package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router) {
	router.Get(constants.Transactions, middlewares.Protected(), ioc.TransactionController.Transactions)
	router.Get(constants.TransactionDetail, middlewares.Protected(), ioc.TransactionController.TransactionById)
	router.Post(constants.Transactions, middlewares.Protected(), ioc.TransactionController.CreateTransaction)
	router.Post(constants.TransactionClone, middlewares.Protected(), ioc.TransactionController.CloneTransaction)

	router.Get(constants.TransactionIdAndItemId, middlewares.Protected(), ioc.TransactionController.TransactionItemById)
	router.Post(constants.TransactionId, middlewares.Protected(), ioc.TransactionController.CreateTransactionItem)
	router.Put(constants.TransactionIdAndItemId, middlewares.Protected(), ioc.TransactionController.UpdateTransactionItem)
	router.Patch(constants.TransactionIdAndItemId, middlewares.Protected(), ioc.TransactionController.MarkAsPaidTransactionItem)
	router.Delete(constants.TransactionIdAndItemId, middlewares.Protected(), ioc.TransactionController.RemoveTransactionItem)
}
