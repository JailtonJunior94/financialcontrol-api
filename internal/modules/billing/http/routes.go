package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router, controller *BillController) {
	router.Get(pkgroutes.Bills, platformsecurity.Protected(), controller.Bills)
	router.Get(pkgroutes.BillDetail, platformsecurity.Protected(), controller.BillById)
	router.Post(pkgroutes.Bills, platformsecurity.Protected(), controller.CreateBill)
	router.Get(pkgroutes.BillsIdAndItemId, platformsecurity.Protected(), controller.BillItemById)
	router.Post(pkgroutes.BillId, platformsecurity.Protected(), controller.CreateBillItem)
	router.Put(pkgroutes.BillsIdAndItemId, platformsecurity.Protected(), controller.UpdateBillItem)
	router.Delete(pkgroutes.BillsIdAndItemId, platformsecurity.Protected(), controller.RemoveBillItem)
}
