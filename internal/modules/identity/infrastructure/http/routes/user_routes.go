package routes

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(router fiber.Router, userHandler *handlers.UserHandler) {
	router.Post(pkgroutes.Users, userHandler.Create)
}
