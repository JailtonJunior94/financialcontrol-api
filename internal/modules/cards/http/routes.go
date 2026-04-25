package http

import (
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AddCardRouter(router fiber.Router, controller *CardController) {
	router.Get(pkgroutes.Cards, platformsecurity.Protected(), controller.Cards)
	router.Get(pkgroutes.CardId, platformsecurity.Protected(), controller.CardById)
	router.Post(pkgroutes.Cards, platformsecurity.Protected(), controller.CreateCard)
	router.Put(pkgroutes.CardId, platformsecurity.Protected(), controller.UpdateCard)
	router.Delete(pkgroutes.CardId, platformsecurity.Protected(), controller.RemoveCard)
}
