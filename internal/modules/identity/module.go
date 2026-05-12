package identity

import (
	"context"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/http/routes"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/repositories/mssql"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

	"github.com/gofiber/fiber/v2"
)

// Deps holds the external dependencies required to build the identity module.
type Deps struct {
	DB          devkitdb.DBTX
	Hasher      interfaces.Hasher
	TokenIssuer interfaces.TokenIssuer
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

// hasherAdapter adapts security.HashAdapter to domain interfaces.Hasher.
type hasherAdapter struct {
	inner platformsecurity.HashAdapter
}

// NewHasherAdapter returns an interfaces.Hasher backed by security.HashAdapter.
func NewHasherAdapter(h platformsecurity.HashAdapter) interfaces.Hasher {
	return &hasherAdapter{inner: h}
}

func (a *hasherAdapter) Hash(plain string) (vos.HashedPassword, error) {
	s, err := a.inner.Hash(plain)
	if err != nil {
		return "", err
	}
	return vos.NewHashedPassword(s)
}

func (a *hasherAdapter) Verify(hashed vos.HashedPassword, plain string) bool {
	return a.inner.Verify(hashed.String(), plain)
}

// tokenIssuerAdapter adapts pkg/jwt.Issuer to domain interfaces.TokenIssuer.
type tokenIssuerAdapter struct {
	inner pkgjwt.Issuer
}

// NewTokenIssuerAdapter returns an interfaces.TokenIssuer backed by pkg/jwt.Issuer.
func NewTokenIssuerAdapter(issuer pkgjwt.Issuer) interfaces.TokenIssuer {
	return &tokenIssuerAdapter{inner: issuer}
}

func (a *tokenIssuerAdapter) Issue(ctx context.Context, userID vos.UserID, email vos.Email) (string, time.Time, error) {
	return a.inner.Issue(ctx, pkgjwt.Identity{
		UserID: userID.String(),
		Email:  email.String(),
	})
}
