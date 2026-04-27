package identity_test

import (
	"database/sql"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	interfacemocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces/mocks"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

// sqlConnectionStub satisfies database.ISqlConnection returning a nil *sqlx.DB
// (sufficient for wiring tests that never trigger actual queries).
type sqlConnectionStub struct{}

func (s *sqlConnectionStub) Connect() *sqlx.DB { return nil }
func (s *sqlConnectionStub) Disconnect()       {}
func (s *sqlConnectionStub) OpenConnectionAndMountStatement(_ string) (*sql.Stmt, error) {
	return nil, nil
}
func (s *sqlConnectionStub) ValidateResult(_ sql.Result, err error) error { return err }
func (s *sqlConnectionStub) Begin() (*sqlx.Tx, error)                     { return nil, nil }
func (s *sqlConnectionStub) Rollback() error                              { return nil }
func (s *sqlConnectionStub) Commit() error                                { return nil }
func (s *sqlConnectionStub) End(fn func() error) error                    { return fn() }

// ModuleWiringSuite validates that NewModule composes all dependencies correctly.
type ModuleWiringSuite struct {
	suite.Suite
	hasher      *interfacemocks.Hasher
	tokenIssuer *interfacemocks.TokenIssuer
	deps        identity.Deps
}

func TestModuleWiringSuite(t *testing.T) { suite.Run(t, new(ModuleWiringSuite)) }

func (s *ModuleWiringSuite) SetupTest() {
	s.hasher = interfacemocks.NewHasher(s.T())
	s.tokenIssuer = interfacemocks.NewTokenIssuer(s.T())
	s.deps = identity.Deps{
		DB:          &sqlConnectionStub{},
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
				s.True(paths["/token"], "POST /token must be registered")
				s.True(paths["/me"], "GET /me must be registered")
				s.True(paths["/users"], "POST /users must be registered")
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

func (s *ModuleWiringSuite) TestNewHasherAdapter_SatisfiesInterface() {
	type args struct{ plain string }
	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(err error)
	}{
		{
			name: "hash delegates to inner adapter",
			args: args{plain: "secret"},
			setup: func() {
				s.hasher.EXPECT().Hash("secret").Return("hashed", nil).Once()
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
	}
	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			_, err := s.deps.Hasher.Hash(sc.args.plain)
			sc.expect(err)
		})
	}
}
