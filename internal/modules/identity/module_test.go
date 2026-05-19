package identity_test

import (
	"context"
	"testing"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports/mocks"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
)

// dbtxStub satisfies devkitdb.DBTX returning zero values
// (sufficient for wiring tests that never trigger actual queries).
type dbtxStub struct{}

func (s *dbtxStub) ExecContext(_ context.Context, _ string, _ ...any) (devkitdb.Result, error) {
	return nil, nil
}
func (s *dbtxStub) QueryContext(_ context.Context, _ string, _ ...any) (devkitdb.Rows, error) {
	return nil, nil
}
func (s *dbtxStub) QueryRowContext(_ context.Context, _ string, _ ...any) devkitdb.Row {
	return nil
}

// ModuleWiringSuite validates that NewModule composes all dependencies correctly.
type ModuleWiringSuite struct {
	suite.Suite
	hasher      *portmocks.Hasher
	tokenIssuer *portmocks.TokenIssuer
	deps        identity.Deps
}

func TestModuleWiringSuite(t *testing.T) { suite.Run(t, new(ModuleWiringSuite)) }

func (s *ModuleWiringSuite) SetupTest() {
	s.hasher = portmocks.NewHasher(s.T())
	s.tokenIssuer = portmocks.NewTokenIssuer(s.T())
	s.deps = identity.Deps{
		DB:          &dbtxStub{},
		Hasher:      s.hasher,
		TokenIssuer: s.tokenIssuer,
	}
}

func (s *ModuleWiringSuite) TestNewModule_PopulatesAllFields() {
	type args struct{ deps identity.Deps }
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(m *identity.Module)
	}{
		{
			name:  "all dependencies provided — module fully wired",
			args:  args{deps: s.deps},
			setup: func() {},
			expect: func(m *identity.Module) {
				s.Require().NotNil(m)
				s.NotNil(m.AuthenticateUser, "AuthenticateUser must be wired")
				s.NotNil(m.GetAuthenticatedUser, "GetAuthenticatedUser must be wired")
				s.NotNil(m.CreateUser, "CreateUser must be wired")
				s.NotNil(m.AuthHandler, "AuthHandler must be wired")
				s.NotNil(m.UserHandler, "UserHandler must be wired")
			},
		},
	}
	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			m := identity.NewModule(sc.args.deps)
			sc.expect(m)
		})
	}
}

func (s *ModuleWiringSuite) TestRegisterHTTP_RegistersRoutes() {
	type args struct{ deps identity.Deps }
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(app *fiber.App)
	}{
		{
			name:  "routes registered without panic",
			args:  args{deps: s.deps},
			setup: func() {},
			expect: func(app *fiber.App) {
				routes := app.GetRoutes()
				paths := make(map[string]bool)
				for _, r := range routes {
					paths[r.Path] = true
				}
				s.True(paths[pkgroutes.Token], "POST /token must be registered")
				s.True(paths[pkgroutes.Me], "GET /me must be registered")
				s.True(paths[pkgroutes.Users], "POST /users must be registered")
			},
		},
	}
	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			m := identity.NewModule(sc.args.deps)
			app := fiber.New()
			noop := func(c *fiber.Ctx) error { return c.Next() }
			m.RegisterHTTP(app, noop)
			sc.expect(app)
		})
	}
}
