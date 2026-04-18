package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Get(routeconstants.Bills, middlewares.Protected(), container.BillController.Bills)
	router.Get(routeconstants.BillDetail, middlewares.Protected(), container.BillController.BillById)
	router.Post(routeconstants.Bills, middlewares.Protected(), container.BillController.CreateBill)
	router.Get(routeconstants.BillsIdAndItemId, middlewares.Protected(), container.BillController.BillItemById)
	router.Post(routeconstants.BillId, middlewares.Protected(), container.BillController.CreateBillItem)
	router.Put(routeconstants.BillsIdAndItemId, middlewares.Protected(), container.BillController.UpdateBillItem)
	router.Delete(routeconstants.BillsIdAndItemId, middlewares.Protected(), container.BillController.RemoveBillItem)
}
