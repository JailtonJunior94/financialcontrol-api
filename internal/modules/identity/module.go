package identity

import (
	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/routes"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/persistence/mssql"

	"github.com/gofiber/fiber/v2"
)

// Deps holds the external dependencies required to build the identity module.
type Deps struct {
	DB          devkitdb.DBTX
	Hasher      ports.Hasher
	TokenIssuer ports.TokenIssuer
}

// Module holds all wired usecases and handlers for the identity domain.
type Module struct {
	AuthenticateUser     usecase.AuthenticateUser
	GetAuthenticatedUser usecase.GetAuthenticatedUser
	CreateUser           usecase.CreateUser
	AuthHandler          *handlers.AuthHandler
	UserHandler          *handlers.UserHandler
}

// NewModule builds the identity module from its external dependencies.
func NewModule(deps Deps) *Module {
	repo := mssqlrepo.NewUserRepository(deps.DB)

	authenticateUser := usecase.NewAuthenticateUser(repo, deps.Hasher, deps.TokenIssuer)
	getAuthenticatedUser := usecase.NewGetAuthenticatedUser(repo)
	createUser := usecase.NewCreateUser(repo, deps.Hasher)

	authHandler := handlers.NewAuthHandler(authenticateUser, getAuthenticatedUser)
	userHandler := handlers.NewUserHandler(createUser)

	return &Module{
		AuthenticateUser:     authenticateUser,
		GetAuthenticatedUser: getAuthenticatedUser,
		CreateUser:           createUser,
		AuthHandler:          authHandler,
		UserHandler:          userHandler,
	}
}

// RegisterHTTP registers all identity routes on the provided router.
// protected is the authentication middleware applied to guarded routes (e.g. /me).
func (m *Module) RegisterHTTP(router fiber.Router, protected fiber.Handler) {
	routes.RegisterAuthRoutes(router, m.AuthHandler, protected)
	routes.RegisterUserRoutes(router, m.UserHandler)
}
