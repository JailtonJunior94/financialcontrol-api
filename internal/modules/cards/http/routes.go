package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddCardRouter(router fiber.Router, controller *CardController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Cards, authmiddleware.Protected(parser), controller.Cards)
	router.Get(pkgroutes.CardId, authmiddleware.Protected(parser), controller.CardById)
	router.Post(pkgroutes.Cards, authmiddleware.Protected(parser), controller.CreateCard)
	router.Put(pkgroutes.CardId, authmiddleware.Protected(parser), controller.UpdateCard)
	router.Delete(pkgroutes.CardId, authmiddleware.Protected(parser), controller.RemoveCard)
}
