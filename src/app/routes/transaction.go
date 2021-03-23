package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router) {
	router.Get("/transactions", middlewares.Protected(), ioc.TransactionController.Transactions)
	router.Get("/transactions/:id", middlewares.Protected(), ioc.TransactionController.TransactionById)
	router.Post("/transactions", middlewares.Protected(), ioc.TransactionController.CreateTransaction)
	router.Get("/transactions/:transactionid/items/:id", middlewares.Protected(), ioc.TransactionController.TransactionItemById)
	router.Post("/transactions/:transactionid", middlewares.Protected(), ioc.TransactionController.CreateTransactionItem)
	router.Put("/transactions/:transactionid/items/:id", middlewares.Protected(), ioc.TransactionController.UpdateTransactionItem)
	router.Delete("/transactions/:transactionid/items/:id", middlewares.Protected(), ioc.TransactionController.RemoveTransactionItem)
}
