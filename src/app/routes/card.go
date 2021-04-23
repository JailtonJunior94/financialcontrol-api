package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/constants"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddCardRouter(router fiber.Router) {
	router.Get(constants.Cards, middlewares.Protected(), ioc.CardController.Cards)
	router.Get(constants.CardId, middlewares.Protected(), ioc.CardController.CardById)
	router.Post(constants.Cards, middlewares.Protected(), ioc.CardController.CreateCard)
}
