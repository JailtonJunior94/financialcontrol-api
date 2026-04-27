package http

import (
	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func AddFlagRouter(router fiber.Router, controller *FlagController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Flags, authmiddleware.Protected(parser), controller.Flags)
}

func AddCategoryRouter(router fiber.Router, controller *CategoryController, parser pkgjwt.Parser) {
	router.Get(pkgroutes.Categories, authmiddleware.Protected(parser), controller.Categories)
}
