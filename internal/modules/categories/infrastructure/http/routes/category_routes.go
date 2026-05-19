package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

func RegisterCategoryRoutes(router fiber.Router, handler *handlers.CategoryHandler, protected fiber.Handler) {
	router.Get(pkgroutes.Categories, protected, handler.List)
	router.Get(pkgroutes.CategoryId, protected, handler.Get)
	router.Post(pkgroutes.Categories, protected, handler.Create)
	router.Put(pkgroutes.CategoryId, protected, handler.Update)
	router.Delete(pkgroutes.CategoryId, protected, handler.Delete)
}
