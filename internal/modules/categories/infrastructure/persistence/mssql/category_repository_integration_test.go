//go:build integration

package mssql_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
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
    [Id]        CHAR(36)      NOT NULL PRIMARY KEY,
    [UserId]    CHAR(36)      NOT NULL,
    [ParentId]  CHAR(36)      NULL,
    [Name]      NVARCHAR(100) NOT NULL,
    [Color]     VARCHAR(32)   NOT NULL,
    [Icon]      VARCHAR(64)   NOT NULL,
    [CreatedAt] DATETIME2     NOT NULL,
    [UpdatedAt] DATETIME2     NOT NULL,
    [DeletedAt] DATETIME2     NULL,
    CONSTRAINT FK_Category_Parent FOREIGN KEY ([ParentId]) REFERENCES dbo.[Category]([Id])
)
`

const createRootUniqueIndex = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Root_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Root_Name_Active
    ON dbo.[Category]([UserId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NULL
`

const createSubUniqueIndex = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Sub_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Sub_Name_Active
    ON dbo.[Category]([UserId], [ParentId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NOT NULL
`

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

type CategoryRepositorySuite struct {
	suite.Suite

	db    *sqlx.DB
	ctx   context.Context
	clock fixedClock
}

func TestCategoryRepositorySuite(t *testing.T) {
	suite.Run(t, new(CategoryRepositorySuite))
}

func (s *CategoryRepositorySuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")
	s.db = db

	for _, ddl := range []string{createCategoryTable, createRootUniqueIndex, createSubUniqueIndex} {
		_, err := s.db.ExecContext(context.Background(), ddl)
		s.Require().NoError(err, "failed to apply ddl")
	}
}

func (s *CategoryRepositorySuite) SetupTest() {
	s.ctx = context.Background()
	s.clock = fixedClock{now: time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)}

	// Children must be deleted before parents because of the self FK.
	_, err := s.db.ExecContext(s.ctx, `DELETE FROM dbo.[Category] WHERE [ParentId] IS NOT NULL`)
	s.Require().NoError(err)
	_, err = s.db.ExecContext(s.ctx, `DELETE FROM dbo.[Category]`)
	s.Require().NoError(err)
}

func (s *CategoryRepositorySuite) newRoot(userID identityvo.UserID, name string) *entities.Category {
	c, err := entities.NewCategory(userID, nil, name, "blue", "wallet", s.clock)
	s.Require().NoError(err)
	return c
}

func (s *CategoryRepositorySuite) newSub(userID identityvo.UserID, parentID vos.CategoryID, name string) *entities.Category {
	pid := parentID
	c, err := entities.NewCategory(userID, &pid, name, "red", "tag", s.clock)
	s.Require().NoError(err)
	return c
}

func (s *CategoryRepositorySuite) TestAddGetByIDList() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Alimentação")
	s.Require().NoError(repo.Add(s.ctx, root))

	got, err := repo.GetByID(s.ctx, userID, root.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(root.ID().String(), got.ID().String())
	s.Equal("Alimentação", got.Name().String())
	s.True(got.IsRoot())
	s.True(got.IsActive())

	items, total, err := repo.List(s.ctx, userID, interfaces.ListFilter{})
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(items, 1)
}

func (s *CategoryRepositorySuite) TestUpdateRenameAndExistsByName() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Lazer")
	s.Require().NoError(repo.Add(s.ctx, root))

	exists, err := repo.ExistsByName(s.ctx, userID, mustName("Lazer"), nil, nil)
	s.Require().NoError(err)
	s.True(exists)

	// Exclude self → false (used by update path before renaming).
	id := root.ID()
	exists, err = repo.ExistsByName(s.ctx, userID, mustName("Lazer"), nil, &id)
	s.Require().NoError(err)
	s.False(exists)

	s.Require().NoError(root.Rename(mustName("Entretenimento"), s.clock))
	s.Require().NoError(repo.Update(s.ctx, root))

	got, err := repo.GetByID(s.ctx, userID, root.ID())
	s.Require().NoError(err)
	s.Equal("Entretenimento", got.Name().String())
}

func (s *CategoryRepositorySuite) TestUniqueNameConflictActive() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	first := s.newRoot(userID, "Casa")
	s.Require().NoError(repo.Add(s.ctx, first))

	dup := s.newRoot(userID, "Casa")
	err := repo.Add(s.ctx, dup)
	s.Require().ErrorIs(err, domain.ErrCategoryNameAlreadyExists)
}

func (s *CategoryRepositorySuite) TestNameReuseAfterSoftDelete() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	first := s.newRoot(userID, "Saúde")
	s.Require().NoError(repo.Add(s.ctx, first))

	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, first.ID(), s.clock.Now()))

	reused := s.newRoot(userID, "Saúde")
	err := repo.Add(s.ctx, reused)
	s.Require().NoError(err, "name should be reusable after soft delete")
}

func (s *CategoryRepositorySuite) TestGetByIDIncludingDeletedReturnsInactiveRow() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Saúde")
	s.Require().NoError(repo.Add(s.ctx, root))
	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, root.ID(), s.clock.Now()))

	got, err := repo.GetByIDIncludingDeleted(s.ctx, userID, root.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.False(got.IsActive())
}

func (s *CategoryRepositorySuite) TestSoftDeleteCascadeAndIdempotency() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Mercado")
	s.Require().NoError(repo.Add(s.ctx, root))

	child1 := s.newSub(userID, root.ID(), "Frutas")
	child2 := s.newSub(userID, root.ID(), "Carnes")
	s.Require().NoError(repo.Add(s.ctx, child1))
	s.Require().NoError(repo.Add(s.ctx, child2))

	deletedAt := time.Date(2026, time.May, 5, 9, 0, 0, 0, time.UTC)
	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, root.ID(), deletedAt))

	for _, id := range []vos.CategoryID{root.ID(), child1.ID(), child2.ID()} {
		_, err := repo.GetByID(s.ctx, userID, id)
		s.Require().ErrorIs(err, domain.ErrCategoryNotFound, "row %s should be soft deleted", id.String())
	}

	// Calling twice must remain a no-op (idempotent).
	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, root.ID(), deletedAt.Add(time.Hour)))

	// And inactive children must not appear in GetActiveChildren.
	children, err := repo.GetActiveChildren(s.ctx, userID, root.ID())
	s.Require().NoError(err)
	s.Empty(children)
}

func (s *CategoryRepositorySuite) TestSoftDeleteCascadeNotFound() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	err := repo.SoftDeleteCascade(s.ctx, userID, vos.NewCategoryID(), s.clock.Now())
	s.Require().ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *CategoryRepositorySuite) TestUpdateDeletedRowReturnsNotFound() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Mercado")
	s.Require().NoError(repo.Add(s.ctx, root))
	s.Require().NoError(repo.SoftDeleteCascade(s.ctx, userID, root.ID(), s.clock.Now()))

	s.Require().NoError(root.Rename(mustName("Mercado Atualizado"), s.clock))
	err := repo.Update(s.ctx, root)
	s.Require().ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *CategoryRepositorySuite) TestCrossUserIsolation() {
	repo := repomssql.NewCategoryRepository(s.db)

	owner := identityvo.NewUserID()
	other := identityvo.NewUserID()

	root := s.newRoot(owner, "Pessoal")
	s.Require().NoError(repo.Add(s.ctx, root))

	got, err := repo.GetByID(s.ctx, other, root.ID())
	s.Require().ErrorIs(err, domain.ErrCategoryNotFound)
	s.Nil(got)

	items, total, err := repo.List(s.ctx, other, interfaces.ListFilter{})
	s.Require().NoError(err)
	s.Equal(int64(0), total)
	s.Empty(items)
}

func (s *CategoryRepositorySuite) TestListFilters() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root1 := s.newRoot(userID, "Transporte")
	root2 := s.newRoot(userID, "Trabalho")
	s.Require().NoError(repo.Add(s.ctx, root1))
	s.Require().NoError(repo.Add(s.ctx, root2))

	sub1 := s.newSub(userID, root1.ID(), "Uber")
	s.Require().NoError(repo.Add(s.ctx, sub1))

	// Only roots
	items, total, err := repo.List(s.ctx, userID, interfaces.ListFilter{OnlyRoots: true})
	s.Require().NoError(err)
	s.Equal(int64(2), total)
	s.Len(items, 2)

	// Only subs
	items, total, err = repo.List(s.ctx, userID, interfaces.ListFilter{OnlySubs: true})
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(items, 1)
	s.Equal("Uber", items[0].Name().String())

	// Subs of root1
	parentID := root1.ID()
	items, total, err = repo.List(s.ctx, userID, interfaces.ListFilter{ParentID: &parentID})
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(items, 1)
	s.Equal("Uber", items[0].Name().String())

	// NameLike filter
	items, total, err = repo.List(s.ctx, userID, interfaces.ListFilter{NameLike: "Trans"})
	s.Require().NoError(err)
	s.Equal(int64(1), total)
	s.Len(items, 1)
	s.Equal("Transporte", items[0].Name().String())

	// Pagination
	items, total, err = repo.List(s.ctx, userID, interfaces.ListFilter{
		Pagination: vos.NewPagination(1, 2),
	})
	s.Require().NoError(err)
	s.Equal(int64(3), total)
	s.Len(items, 2)
}

func (s *CategoryRepositorySuite) TestGetActiveChildren() {
	repo := repomssql.NewCategoryRepository(s.db)
	userID := identityvo.NewUserID()

	root := s.newRoot(userID, "Contas")
	s.Require().NoError(repo.Add(s.ctx, root))

	c1 := s.newSub(userID, root.ID(), "Luz")
	c2 := s.newSub(userID, root.ID(), "Água")
	s.Require().NoError(repo.Add(s.ctx, c1))
	s.Require().NoError(repo.Add(s.ctx, c2))

	children, err := repo.GetActiveChildren(s.ctx, userID, root.ID())
	s.Require().NoError(err)
	s.Len(children, 2)
}

func mustName(s string) vos.CategoryName {
	n, err := vos.NewCategoryName(s)
	if err != nil {
		panic(err)
	}
	return n
}
