package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddBillRouter(router fiber.Router, controller *BillController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Bills, authmiddleware.Protected(parser), controller.Bills)
	router.Get(pkgroutes.BillDetail, authmiddleware.Protected(parser), controller.BillById)
	router.Post(pkgroutes.Bills, authmiddleware.Protected(parser), controller.CreateBill)
	router.Get(pkgroutes.BillsIdAndItemId, authmiddleware.Protected(parser), controller.BillItemById)
	router.Post(pkgroutes.BillId, authmiddleware.Protected(parser), controller.CreateBillItem)
	router.Put(pkgroutes.BillsIdAndItemId, authmiddleware.Protected(parser), controller.UpdateBillItem)
	router.Delete(pkgroutes.BillsIdAndItemId, authmiddleware.Protected(parser), controller.RemoveBillItem)
}
