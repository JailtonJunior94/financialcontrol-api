package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router, controller *BillController) {
	router.Get(routeconstants.Bills, middlewares.Protected(), controller.Bills)
	router.Get(routeconstants.BillDetail, middlewares.Protected(), controller.BillById)
	router.Post(routeconstants.Bills, middlewares.Protected(), controller.CreateBill)
	router.Get(routeconstants.BillsIdAndItemId, middlewares.Protected(), controller.BillItemById)
	router.Post(routeconstants.BillId, middlewares.Protected(), controller.CreateBillItem)
	router.Put(routeconstants.BillsIdAndItemId, middlewares.Protected(), controller.UpdateBillItem)
	router.Delete(routeconstants.BillsIdAndItemId, middlewares.Protected(), controller.RemoveBillItem)
}
