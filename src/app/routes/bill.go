package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router) {
	router.Get("/bills", middlewares.Protected(), ioc.BillController.Bills)
	router.Get("/bills/:id", middlewares.Protected(), ioc.BillController.BillById)
	router.Post("/bills", middlewares.Protected(), ioc.BillController.CreateBill)
	router.Get("/bills/:billid/items/:id", middlewares.Protected(), ioc.BillController.BillItemById)
	router.Post("/bills/:billid", middlewares.Protected(), ioc.BillController.CreateBillItem)
	router.Put("/bills/:billid/items/:id", middlewares.Protected(), ioc.BillController.UpdateBillItem)
	router.Delete("/bills/:billid/items/:id", middlewares.Protected(), ioc.BillController.RemoveBillItem)
}
