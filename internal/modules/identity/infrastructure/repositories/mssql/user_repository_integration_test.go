//go:build integration

package mssql_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"

	repomssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure/repositories/mssql"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

const createUserTable = `
IF NOT EXISTS (
    SELECT * FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'User'
)
CREATE TABLE dbo.[User] (
    [Id]        CHAR(36)       NOT NULL PRIMARY KEY,
    [Name]      NVARCHAR(255)  NOT NULL,
    [Email]     NVARCHAR(255)  NOT NULL UNIQUE,
    [Password]  NVARCHAR(255)  NOT NULL,
    [CreatedAt] DATETIME2      NOT NULL,
    [UpdatedAt] DATETIME2      NOT NULL,
    [Active]    BIT            NOT NULL DEFAULT 1
)
`

type UserRepositorySuite struct {
	suite.Suite

	db      *sqlx.DB
	ctx     context.Context
	cleanup func()
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (s *UserRepositorySuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")
	s.db = db

	_, err = s.db.ExecContext(context.Background(), createUserTable)
	s.Require().NoError(err, "failed to create user table")
}

func (s *UserRepositorySuite) SetupTest() {
	s.ctx = context.Background()

	_, cleanup, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")
	s.cleanup = cleanup

	cleanup()
}

func (s *UserRepositorySuite) TearDownTest() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *UserRepositorySuite) newUser(name, email, password string) *entities.User {
	e, err := vos.NewEmail(email)
	s.Require().NoError(err)
	pwd, err := vos.NewHashedPassword(password)
	s.Require().NoError(err)
	u, err := entities.NewUser(name, e, pwd)
	s.Require().NoError(err)
	return u
}

func (s *UserRepositorySuite) TestAdd() {
	repo := repomssql.NewUserRepository(s.db)

	scenarios := []struct {
		name   string
		setup  func()
		user   func() *entities.User
		expect func(err error)
	}{
		{
			name:  "insere usuário com sucesso",
			setup: func() {},
			user: func() *entities.User {
				return s.newUser("Alice", "alice-add@example.com", "$2a$10$hash1")
			},
			expect: func(err error) {
				s.Require().NoError(err)
			},
		},
		{
			name: "conflito de e-mail retorna erro",
			setup: func() {
				u := s.newUser("Alice", "alice-conflict@example.com", "$2a$10$hash2")
				s.Require().NoError(repo.Add(s.ctx, u))
			},
			user: func() *entities.User {
				return s.newUser("Alice2", "alice-conflict@example.com", "$2a$10$other")
			},
			expect: func(err error) {
				s.Require().Error(err, "duplicate email must return error")
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			err := repo.Add(s.ctx, sc.user())
			sc.expect(err)
		})
	}
}

func (s *UserRepositorySuite) TestGetByEmail() {
	repo := repomssql.NewUserRepository(s.db)

	alice := s.newUser("Alice", "alice-getemail@example.com", "$2a$10$hash")
	s.Require().NoError(repo.Add(s.ctx, alice))

	scenarios := []struct {
		name   string
		setup  func()
		email  func() vos.Email
		expect func(u *entities.User, err error)
	}{
		{
			name:  "retorna usuário existente por e-mail",
			setup: func() {},
			email: func() vos.Email { return alice.Email() },
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(u)
				s.Equal(alice.ID().String(), u.ID().String())
				s.Equal(alice.Name(), u.Name())
				s.Equal(alice.Email(), u.Email())
			},
		},
		{
			name:  "retorna nil para e-mail não encontrado",
			setup: func() {},
			email: func() vos.Email {
				e, _ := vos.NewEmail("notfound@example.com")
				return e
			},
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Nil(u)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			u, err := repo.GetByEmail(s.ctx, sc.email())
			sc.expect(u, err)
		})
	}
}

func (s *UserRepositorySuite) TestGetByID() {
	repo := repomssql.NewUserRepository(s.db)

	bob := s.newUser("Bob", "bob-getid@example.com", "$2a$10$hash")
	s.Require().NoError(repo.Add(s.ctx, bob))

	scenarios := []struct {
		name   string
		setup  func()
		id     func() vos.UserID
		expect func(u *entities.User, err error)
	}{
		{
			name:  "retorna usuário existente por ID",
			setup: func() {},
			id:    func() vos.UserID { return bob.ID() },
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(u)
				s.Equal(bob.ID().String(), u.ID().String())
				s.Equal(bob.Name(), u.Name())
				s.Equal(bob.Email(), u.Email())
			},
		},
		{
			name:  "retorna nil para ID não encontrado",
			setup: func() {},
			id:    func() vos.UserID { return vos.NewUserID() },
			expect: func(u *entities.User, err error) {
				s.Require().NoError(err)
				s.Nil(u)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			u, err := repo.GetByID(s.ctx, sc.id())
			sc.expect(u, err)
		})
	}
}
