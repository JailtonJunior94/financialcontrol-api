package routes

import (
	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	routeconstants "github.com/jailtonjunior94/financialcontrol-api/internal/http/constants"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AddCardRouter(router fiber.Router, container *bootstrapcontainer.Container) {
	router.Get(routeconstants.Cards, middlewares.Protected(), container.CardController.Cards)
	router.Get(routeconstants.CardId, middlewares.Protected(), container.CardController.CardById)
	router.Post(routeconstants.Cards, middlewares.Protected(), container.CardController.CreateCard)
	router.Put(routeconstants.CardId, middlewares.Protected(), container.CardController.UpdateCard)
	router.Delete(routeconstants.CardId, middlewares.Protected(), container.CardController.RemoveCard)
}
