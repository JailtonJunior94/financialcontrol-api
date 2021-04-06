package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router) {
	router.Get("/bills", middlewares.Protected(), ioc.BillController.Bills)
	router.Post("/bills", middlewares.Protected(), ioc.BillController.CreateBill)
}
