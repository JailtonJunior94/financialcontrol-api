package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Get(routeconstants.Transactions, middlewares.Protected(), container.TransactionController.Transactions)
	router.Get(routeconstants.TransactionDetail, middlewares.Protected(), container.TransactionController.TransactionById)
	router.Post(routeconstants.Transactions, middlewares.Protected(), container.TransactionController.CreateTransaction)
	router.Post(routeconstants.TransactionClone, middlewares.Protected(), container.TransactionController.CloneTransaction)
	router.Get(routeconstants.TransactionIdAndItemId, middlewares.Protected(), container.TransactionController.TransactionItemById)
	router.Post(routeconstants.TransactionId, middlewares.Protected(), container.TransactionController.CreateTransactionItem)
	router.Put(routeconstants.TransactionIdAndItemId, middlewares.Protected(), container.TransactionController.UpdateTransactionItem)
	router.Patch(routeconstants.TransactionIdAndItemId, middlewares.Protected(), container.TransactionController.MarkAsPaidTransactionItem)
	router.Delete(routeconstants.TransactionIdAndItemId, middlewares.Protected(), container.TransactionController.RemoveTransactionItem)
}
