package categories

import (
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/routes"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/persistence/mssql"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

// Deps holds the external dependencies required to build the categories module.
type Deps struct {
	DB        devkitdb.DBTX
	JwtParser pkgjwt.Parser
}

// Module holds all wired use cases and the HTTP handler for the categories domain.
type Module struct {
	CreateCategory  usecase.CreateCategory
	UpdateCategory  usecase.UpdateCategory
	DeleteCategory  usecase.DeleteCategory
	GetCategory     usecase.GetCategory
	ListCategories  usecase.ListCategories
	CategoryHandler *handlers.CategoryHandler
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// NewModule builds the categories module from its external dependencies.
func NewModule(deps Deps) *Module {
	repo := mssqlrepo.NewCategoryRepository(deps.DB)

	clock := systemClock{}
	uniqueness := services.NewCategoryUniquenessService(repo)
	deletion := services.NewCategoryDeletionService(repo, clock)

	create := usecase.NewCreateCategory(repo, uniqueness, clock)
	update := usecase.NewUpdateCategory(repo, uniqueness, clock)
	del := usecase.NewDeleteCategory(deletion)
	get := usecase.NewGetCategory(repo)
	list := usecase.NewListCategories(repo)

	handler := handlers.NewCategoryHandler(list, get, create, update, del)

	return &Module{
		CreateCategory:  create,
		UpdateCategory:  update,
		DeleteCategory:  del,
		GetCategory:     get,
		ListCategories:  list,
		CategoryHandler: handler,
	}
}

// RegisterHTTP registers all categories routes on the provided router.
func (m *Module) RegisterHTTP(router fiber.Router, parser pkgjwt.Parser) {
	routes.RegisterCategoryRoutes(router, m.CategoryHandler, parser)
}
