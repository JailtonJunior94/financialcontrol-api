//go:build integration

package mssql_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	repomssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

type FlagRepositorySuite struct {
	suite.Suite
}

func TestFlagRepositorySuite(t *testing.T) {
	suite.Run(t, new(FlagRepositorySuite))
}

func (s *FlagRepositorySuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")

	// Ensure Flag table exists (may already be created by CardRepositorySuite)
	_, err = db.ExecContext(s.T().Context(), createFlagTable)
	s.Require().NoError(err, "failed to create flag table")

	// Ensure seed flag exists
	_, err = db.ExecContext(s.T().Context(), insertSeedFlag)
	s.Require().NoError(err, "failed to seed flag")
}

func (s *FlagRepositorySuite) TestListReturnsSeed() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)

	repo := repomssql.NewFlagRepository(db)
	flags, err := repo.List(s.T().Context())
	s.Require().NoError(err)
	s.NotEmpty(flags, "List should return at least the seed flag")

	var found bool
	for _, f := range flags {
		if f.ID().String() == "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
			found = true
			s.Equal("Visa", f.Name())
			s.True(f.Active())
		}
	}
	s.True(found, "seed flag Visa must be present in List result")
}

func (s *FlagRepositorySuite) TestListAlwaysReturnsFullCatalog() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)

	_, err = db.ExecContext(s.T().Context(), `
IF NOT EXISTS (SELECT 1 FROM dbo.[Flag] WHERE [Id] = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb')
INSERT INTO dbo.[Flag] ([Id], [Name], [Active])
VALUES ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'Mastercard', 1)
`)
	s.Require().NoError(err)

	repo := repomssql.NewFlagRepository(db)

	allFlags, err := repo.List(s.T().Context())
	s.Require().NoError(err)
	s.GreaterOrEqual(len(allFlags), 2)
}

func (s *FlagRepositorySuite) TestExists() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)

	repo := repomssql.NewFlagRepository(db)

	existingID, err := vos.ParseFlagID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	s.Require().NoError(err)

	nonExistentID, err := vos.ParseFlagID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	s.Require().NoError(err)

	scenarios := []struct {
		name   string
		id     vos.FlagID
		expect bool
	}{
		{
			name:   "ID existente retorna true",
			id:     existingID,
			expect: true,
		},
		{
			name:   "ID inexistente retorna false",
			id:     nonExistentID,
			expect: false,
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			ok, err := repo.Exists(s.T().Context(), sc.id)
			s.Require().NoError(err)
			s.Equal(sc.expect, ok)
		})
	}
}
