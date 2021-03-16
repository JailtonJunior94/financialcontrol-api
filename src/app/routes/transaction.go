package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddTransactionRouter(router fiber.Router) {
	router.Post("/transactions", middlewares.Protected(), ioc.TransactionController.CreateTransaction)
	router.Post("/transactions/:id", middlewares.Protected(), ioc.TransactionController.CreateTransactionItem)
	router.Get("/transactions/:id", middlewares.Protected(), ioc.TransactionController.TransactionById)
}
