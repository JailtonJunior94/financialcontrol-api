package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler, protected fiber.Handler) {
	router.Post(pkgroutes.Token, authHandler.Authenticate)
	router.Get("/me", protected, authHandler.Me)
}
