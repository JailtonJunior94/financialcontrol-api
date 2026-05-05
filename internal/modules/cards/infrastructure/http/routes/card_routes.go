package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func RegisterCardRoutes(router fiber.Router, card *handlers.CardHandler, flag *handlers.FlagHandler, parser pkgjwt.Parser) {
	protected := authmiddleware.Protected(parser)

	router.Get(pkgroutes.Cards, protected, card.List)
	router.Get(pkgroutes.CardFlags, protected, flag.List)
	router.Get(pkgroutes.CardId, protected, card.Get)
	router.Post(pkgroutes.Cards, protected, card.Create)
	router.Put(pkgroutes.CardId, protected, card.Update)
	router.Delete(pkgroutes.CardId, protected, card.Deactivate)
}
