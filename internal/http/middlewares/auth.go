package middlewares

import (
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/internal/platform/security"

	"github.com/gofiber/fiber/v2"
)

func Protected() fiber.Handler {
	return platformsecurity.Protected()
}
