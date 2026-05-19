//go:build integration

package mssql_test

import (
	"context"
	"strings"
	"testing"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/stretchr/testify/suite"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	repomssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/persistence/mssql"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const createCategoryTable = `
IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'Category'
)
CREATE TABLE dbo.[Category] (
    [Id]        uniqueidentifier NOT NULL PRIMARY KEY,
    [Name]      varchar(100)     NOT NULL,
    [Sequence]  int              NOT NULL,
    [CreatedAt] datetime2        NOT NULL,
    [UpdatedAt] datetime2        NOT NULL,
    [Active]    bit              NOT NULL
)
`

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

type CategoryRepositorySuite struct {
	suite.Suite

	mgr   devkitmgr.Manager
	dbtx  devkitdb.DBTX
	ctx   context.Context
	clock fixedClock
}

func TestCategoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(CategoryRepositorySuite))
}

func (s *CategoryRepositorySuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")
	_, err = db.ExecContext(context.Background(), createCategoryTable)
	s.Require().NoError(err, "failed to apply ddl")

	mgr, _, err := dbmssql.GetSharedTestManager()
	s.Require().NoError(err, "failed to get shared test manager")
	s.mgr = mgr
}

func (s *CategoryRepositorySuite) SetupTest() {
	s.ctx = context.Background()
	s.dbtx = s.mgr.DBTX(s.ctx)
	s.clock = fixedClock{now: time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)}

	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)
	_, err = db.ExecContext(s.ctx, `DELETE FROM dbo.[Category]`)
	s.Require().NoError(err)
}

func (s *CategoryRepositorySuite) newCategory(name string, sequence int) *entities.Category {
	c, err := entities.NewCategory(name, sequence, s.clock)
	s.Require().NoError(err)
	return c
}

func (s *CategoryRepositorySuite) TestAddGetByIDList() {
	repo := repomssql.NewCategoryRepository(s.dbtx)
	userID := identityvo.NewUserID()

	category := s.newCategory("Alimentação", 5)
	s.Require().NoError(repo.Add(s.ctx, category))

	got, err := repo.GetByID(s.ctx, userID, category.ID())
	s.Require().NoError(err)
	s.True(strings.EqualFold(category.ID().String(), got.ID().String()))
	s.Equal("Alimentação", got.Name().String())
	s.Equal(5, got.Sequence())
	s.True(got.IsActive())

	items, total, err := repo.List(s.ctx, userID, ports.ListFilter{})
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(items, 1)
}

func (s *CategoryRepositorySuite) TestUpdateExistsByNameAndNameReuseAfterDeactivate() {
	repo := repomssql.NewCategoryRepository(s.dbtx)
	userID := identityvo.NewUserID()

	category := s.newCategory("Lazer", 1)
	s.Require().NoError(repo.Add(s.ctx, category))

	exists, err := repo.ExistsByName(s.ctx, userID, mustName("Lazer"), nil, nil)
	s.Require().NoError(err)
	s.True(exists)

	id := category.ID()
	exists, err = repo.ExistsByName(s.ctx, userID, mustName("Lazer"), nil, &id)
	s.Require().NoError(err)
	s.False(exists)

	s.Require().NoError(category.Update(mustName("Entretenimento"), 9, s.clock))
	s.Require().NoError(repo.Update(s.ctx, category))

	got, err := repo.GetByID(s.ctx, userID, category.ID())
	s.Require().NoError(err)
	s.Equal("Entretenimento", got.Name().String())
	s.Equal(9, got.Sequence())

	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, category.ID(), s.clock.Now()))

	reused := s.newCategory("Entretenimento", 10)
	s.Require().NoError(repo.Add(s.ctx, reused), "name should be reusable after deactivation")
}

func (s *CategoryRepositorySuite) TestGetByIDIncludingDeletedAndDeactivate() {
	repo := repomssql.NewCategoryRepository(s.dbtx)
	userID := identityvo.NewUserID()

	category := s.newCategory("Saúde", 2)
	s.Require().NoError(repo.Add(s.ctx, category))
	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, category.ID(), s.clock.Now()))

	_, err := repo.GetByID(s.ctx, userID, category.ID())
	s.Require().ErrorIs(err, domain.ErrCategoryNotFound)

	got, err := repo.GetByIDIncludingDeleted(s.ctx, userID, category.ID())
	s.Require().NoError(err)
	s.False(got.IsActive())

	children, err := repo.GetActiveChildren(s.ctx, userID, category.ID())
	s.Require().NoError(err)
	s.Empty(children)
}

func (s *CategoryRepositorySuite) TestSoftDeleteCascadeNotFoundAndListFilters() {
	repo := repomssql.NewCategoryRepository(s.dbtx)
	userID := identityvo.NewUserID()

	err := repo.SoftDeleteCascade(s.ctx, userID, vos.NewCategoryID(), s.clock.Now())
	s.Require().ErrorIs(err, domain.ErrCategoryNotFound)

	first := s.newCategory("Transporte", 2)
	second := s.newCategory("Trabalho", 1)
	third := s.newCategory("Casa", 3)
	s.Require().NoError(repo.Add(s.ctx, first))
	s.Require().NoError(repo.Add(s.ctx, second))
	s.Require().NoError(repo.Add(s.ctx, third))

	items, total, err := repo.List(s.ctx, userID, ports.ListFilter{NameLike: "Tra"})
	s.Require().NoError(err)
	s.Equal(int64(2), total)
	s.Len(items, 2)
	s.Equal("Trabalho", items[0].Name().String())

	items, total, err = repo.List(s.ctx, userID, ports.ListFilter{
		Pagination: vos.NewPagination(1, 2),
	})
	s.Require().NoError(err)
	s.Equal(int64(3), total)
	s.Len(items, 2)
}

func mustName(s string) vos.CategoryName {
	n, err := vos.NewCategoryName(s)
	if err != nil {
		panic(err)
	}
	return n
}
