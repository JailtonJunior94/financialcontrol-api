package http

import (
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddCardRouter(router fiber.Router, controller *CardController) {
	router.Get(routeconstants.Cards, middlewares.Protected(), controller.Cards)
	router.Get(routeconstants.CardId, middlewares.Protected(), controller.CardById)
	router.Post(routeconstants.Cards, middlewares.Protected(), controller.CreateCard)
	router.Put(routeconstants.CardId, middlewares.Protected(), controller.UpdateCard)
	router.Delete(routeconstants.CardId, middlewares.Protected(), controller.RemoveCard)
}
