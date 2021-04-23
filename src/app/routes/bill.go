package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router) {
	router.Get(constants.Bills, middlewares.Protected(), ioc.BillController.Bills)
	router.Get(constants.BillDetail, middlewares.Protected(), ioc.BillController.BillById)
	router.Post(constants.Bills, middlewares.Protected(), ioc.BillController.CreateBill)
	router.Get(constants.BillsIdAndItemId, middlewares.Protected(), ioc.BillController.BillItemById)
	router.Post(constants.BillId, middlewares.Protected(), ioc.BillController.CreateBillItem)
	router.Put(constants.BillsIdAndItemId, middlewares.Protected(), ioc.BillController.UpdateBillItem)
	router.Delete(constants.BillsIdAndItemId, middlewares.Protected(), ioc.BillController.RemoveBillItem)
}
