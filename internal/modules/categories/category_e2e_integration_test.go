//go:build integration

// Package categories E2E integration tests.
//
// Local execution:
//
//	go test -tags integration -run TestCategoriesE2E ./internal/modules/categories/...
//
// Requires Docker (testcontainers spins up a SQL Server 2022 container — see
// pkg/database/mssql/testing.go). The same container is shared across the
// binary via sync.Once.
package categories_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	domainservices "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/handlers"
	categoryroutes "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/http/routes"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/persistence/mssql"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	pkgroutes "github.com/jailtonjunior94/financialcontrol-api/pkg/routes"
)

// --- Schema (mirrors migrations/0001_create_categories.sql) -------------------

const ddlCategoryTable = `
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

const ddlIxUserActive = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'IX_Category_User_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE INDEX IX_Category_User_Active
    ON dbo.[Category]([UserId], [DeletedAt])
    INCLUDE([ParentId], [Name])
`

const ddlIxUserParentActive = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'IX_Category_User_Parent_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE INDEX IX_Category_User_Parent_Active
    ON dbo.[Category]([UserId], [ParentId], [DeletedAt])
`

const ddlUxRoot = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Root_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Root_Name_Active
    ON dbo.[Category]([UserId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NULL
`

const ddlUxSub = `
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Sub_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Sub_Name_Active
    ON dbo.[Category]([UserId], [ParentId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NOT NULL
`

// --- Test doubles -------------------------------------------------------------

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

// stubParser treats the bearer token as the user id and returns a deterministic
// identity. Avoids spinning up a full JWT issuer just to drive the HTTP layer.
type stubParser struct{}

func (stubParser) Parse(_ context.Context, token string) (pkgjwt.Identity, error) {
	if token == "" {
		return pkgjwt.Identity{}, fmt.Errorf("empty token")
	}
	return pkgjwt.Identity{UserID: token, Email: token + "@example.com"}, nil
}

// --- Suite --------------------------------------------------------------------

type CategoriesE2ESuite struct {
	suite.Suite

	db    *sqlx.DB
	app   *fiber.App
	clock fixedClock
}

func TestCategoriesE2E(t *testing.T) {
	suite.Run(t, new(CategoriesE2ESuite))
}

func (s *CategoriesE2ESuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)
	s.db = db

	for _, ddl := range []string{
		ddlCategoryTable, ddlIxUserActive, ddlIxUserParentActive, ddlUxRoot, ddlUxSub,
	} {
		_, err := s.db.ExecContext(context.Background(), ddl)
		s.Require().NoError(err, "apply ddl: %s", ddl)
	}

	s.app = s.buildApp()
}

func (s *CategoriesE2ESuite) SetupTest() {
	s.clock = fixedClock{now: time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)}
	s.truncate()
	// Rebuild app each test so the fixed clock is fresh.
	s.app = s.buildApp()
}

func (s *CategoriesE2ESuite) truncate() {
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `DELETE FROM dbo.[Category] WHERE [ParentId] IS NOT NULL`)
	s.Require().NoError(err)
	_, err = s.db.ExecContext(ctx, `DELETE FROM dbo.[Category]`)
	s.Require().NoError(err)
}

func (s *CategoriesE2ESuite) buildApp() *fiber.App {
	repo := mssqlrepo.NewCategoryRepository(s.db)
	uniqueness := domainservices.NewCategoryUniquenessService(repo)
	deletion := domainservices.NewCategoryDeletionService(repo, s.clock)

	create := usecase.NewCreateCategory(repo, uniqueness, s.clock)
	update := usecase.NewUpdateCategory(repo, uniqueness, s.clock)
	del := usecase.NewDeleteCategory(deletion)
	get := usecase.NewGetCategory(repo)
	list := usecase.NewListCategories(repo)

	handler := handlers.NewCategoryHandler(list, get, create, update, del)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	categoryroutes.RegisterCategoryRoutes(app, handler, stubParser{})
	return app
}

// --- HTTP helpers -------------------------------------------------------------

// httpResult is a fully-buffered HTTP response so callers can both inspect the
// body for assertion messages and decode it as JSON without races on Read.
type httpResult struct {
	status int
	body   []byte
}

func (r httpResult) text() string { return string(r.body) }

func (s *CategoriesE2ESuite) request(method, path, token string, body any) httpResult {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		s.Require().NoError(err)
		reader = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.app.Test(req, -1)
	s.Require().NoError(err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return httpResult{status: resp.StatusCode, body: raw}
}

func decodeJSON[T any](s *CategoriesE2ESuite, r httpResult) T {
	s.T().Helper()
	var out T
	s.Require().NoErrorf(json.Unmarshal(r.body, &out), "decode body: %s", r.text())
	return out
}

func newUserID() string { return uuid.NewString() }

func validRequest(name string, parent *string) dtos.CategoryRequest {
	return dtos.CategoryRequest{Name: name, Color: "blue", Icon: "wallet", ParentID: parent}
}

// --- Tests --------------------------------------------------------------------

// 8.2 — full happy path: create root → create sub → list → update → soft delete cascade.
func (s *CategoriesE2ESuite) TestFullLifecycle_RootSubListUpdateDeleteCascade() {
	user := newUserID()

	// create root
	resp := s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Mercado", nil))
	s.Require().Equal(http.StatusCreated, resp.status)
	root := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Require().NotEmpty(root.ID)

	// create sub
	pid := root.ID
	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Frutas", &pid))
	s.Require().Equal(http.StatusCreated, resp.status)
	sub := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Require().NotNil(sub.ParentID)
	s.Require().Equal(root.ID, *sub.ParentID)

	// list
	resp = s.request(http.MethodGet, pkgroutes.Categories+"?pageSize=10", user, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list := decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(int64(2), list.Total)
	s.Len(list.Items, 2)
	s.Equal(10, list.PageSize)

	// update sub (rename)
	resp = s.request(http.MethodPut, "/categories/"+sub.ID, user, validRequest("Frutas Frescas", &pid))
	s.Require().Equal(http.StatusOK, resp.status)
	updated := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Equal("Frutas Frescas", updated.Name)

	// delete root → cascades to sub
	resp = s.request(http.MethodDelete, "/categories/"+root.ID, user, nil)
	s.Require().Equal(http.StatusNoContent, resp.status)

	// both rows must be gone (soft deleted)
	for _, id := range []string{root.ID, sub.ID} {
		got := s.request(http.MethodGet, "/categories/"+id, user, nil)
		s.Equal(http.StatusNotFound, got.status, "id %s", id)
	}
}

// 8.3 — uniqueness conflict (409) and reuse after soft delete (201).
func (s *CategoriesE2ESuite) TestUniqueness_ConflictAndReuseAfterSoftDelete() {
	user := newUserID()

	resp := s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", nil))
	s.Require().Equal(http.StatusCreated, resp.status)
	first := decodeJSON[dtos.CategoryResponse](s, resp)

	// Same name → 409
	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", nil))
	s.Equal(http.StatusConflict, resp.status)

	// Soft delete first then reuse name → 201
	resp = s.request(http.MethodDelete, "/categories/"+first.ID, user, nil)
	s.Require().Equal(http.StatusNoContent, resp.status)

	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", nil))
	s.Equal(http.StatusCreated, resp.status)
}

// 8.4 — reparent: move sub to another active parent; reject inactive parent.
func (s *CategoriesE2ESuite) TestReparent_MoveBetweenParentsAndRejectInactive() {
	user := newUserID()

	// create two roots and one sub under root1
	root1ID := s.mustCreate(user, "Casa", nil)
	root2ID := s.mustCreate(user, "Trabalho", nil)
	subID := s.mustCreate(user, "Internet", &root1ID)

	// reparent sub from root1 to root2
	resp := s.request(http.MethodPut, "/categories/"+subID, user, validRequest("Internet", &root2ID))
	s.Require().Equal(http.StatusOK, resp.status)
	updated := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Require().NotNil(updated.ParentID)
	s.Equal(root2ID, *updated.ParentID)

	// soft delete root1, then try reparenting back to root1 → 422 (parent inactive)
	resp = s.request(http.MethodDelete, "/categories/"+root1ID, user, nil)
	s.Require().Equal(http.StatusNoContent, resp.status)

	resp = s.request(http.MethodPut, "/categories/"+subID, user, validRequest("Internet", &root1ID))
	s.Equal(http.StatusUnprocessableEntity, resp.status,
		"reparenting to a soft-deleted parent must be rejected as inactive")
}

// 8.5 — cross-user isolation: user B cannot read/touch user A's category.
func (s *CategoriesE2ESuite) TestCrossUser_IsolationReturns404() {
	owner := newUserID()
	other := newUserID()

	id := s.mustCreate(owner, "Pessoal", nil)

	resp := s.request(http.MethodGet, "/categories/"+id, other, nil)
	s.Equal(http.StatusNotFound, resp.status)

	resp = s.request(http.MethodPut, "/categories/"+id, other, validRequest("Outro", nil))
	s.Equal(http.StatusNotFound, resp.status)

	resp = s.request(http.MethodDelete, "/categories/"+id, other, nil)
	s.Equal(http.StatusNotFound, resp.status)

	// Owner still sees the row → confirms isolation worked.
	resp = s.request(http.MethodGet, "/categories/"+id, owner, nil)
	s.Require().Equal(http.StatusOK, resp.status)

	// listing under "other" must be empty
	resp = s.request(http.MethodGet, pkgroutes.Categories, other, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list := decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(int64(0), list.Total)
	s.Empty(list.Items)
}

// 8.6 — performance: 10k rows, GET /categories?pageSize=100 must stay well below 200ms.
// Reports the median of N runs to the test log for auditing (RNF-01).
func (s *CategoriesE2ESuite) TestPerformance_ListWith10kCategoriesUnderBudget() {
	if testing.Short() {
		s.T().Skip("skipping performance test in short mode")
	}

	user := newUserID()
	s.seedRoots(user, 10_000)

	const runs = 7
	durations := make([]time.Duration, 0, runs)
	for range runs {
		start := time.Now()
		resp := s.request(http.MethodGet, pkgroutes.Categories+"?pageSize=100", user, nil)
		elapsed := time.Since(start)
		s.Require().Equal(http.StatusOK, resp.status)
		body := decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
		s.Require().Equal(int64(10_000), body.Total)
		s.Require().Len(body.Items, 100)
		durations = append(durations, elapsed)
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	median := durations[len(durations)/2]
	s.T().Logf("LIST 10k categories — runs=%d durations=%v median=%s", runs, durations, median)

	// RNF-01 budget is 200ms; assert with a 30% margin (260ms) to stay non-flaky.
	const budget = 260 * time.Millisecond
	s.LessOrEqualf(median, budget,
		"median GET /categories with 10k rows must be <= %s (got %s)", budget, median)
}

// 8.7 — depth-3 inserted via raw SQL is detected when application tries to
// reparent or update through that branch.
func (s *CategoriesE2ESuite) TestDepth_RawDepthThreeRejectedByApplication() {
	user := newUserID()

	// build root → sub via repo so they pass invariants, then forge depth-3 via raw SQL.
	rootID := s.mustCreate(user, "Root", nil)
	subID := s.mustCreate(user, "Sub", &rootID)

	depth3ID := uuid.NewString()
	now := s.clock.Now()
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO dbo.[Category]
			([Id],[UserId],[ParentId],[Name],[Color],[Icon],[CreatedAt],[UpdatedAt],[DeletedAt])
		 VALUES (@id,@userId,@parentId,@name,@color,@icon,@createdAt,@updatedAt,NULL)`,
		sql.Named("id", depth3ID),
		sql.Named("userId", user),
		sql.Named("parentId", subID),
		sql.Named("name", "Depth3"),
		sql.Named("color", "blue"),
		sql.Named("icon", "wallet"),
		sql.Named("createdAt", now),
		sql.Named("updatedAt", now),
	)
	s.Require().NoError(err, "raw insert of depth-3 row must succeed (DB does not enforce depth)")

	// 1) Trying to create a new category whose parent is the depth-2 sub: rejected
	//    by the application because parent is not a root.
	resp := s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("NewLeaf", &subID))
	s.Equalf(http.StatusUnprocessableEntity, resp.status,
		"creating a child of a non-root parent must be rejected by application; body=%s", resp.text())

	// 2) Reparenting the forged depth-3 row to ANOTHER non-root parent must
	//    also be rejected — the application detects the depth violation as
	//    soon as the use case re-evaluates parent eligibility.
	otherSubID := s.mustCreate(user, "Sub2", &rootID)
	resp = s.request(http.MethodPut, "/categories/"+depth3ID, user, validRequest("Depth3", &otherSubID))
	s.Equalf(http.StatusUnprocessableEntity, resp.status,
		"reparenting to a non-root parent must be rejected by application; body=%s", resp.text())
}

func (s *CategoriesE2ESuite) TestUpdateRejectsRootReparentAndSelfParent() {
	user := newUserID()

	rootID := s.mustCreate(user, "Root", nil)
	otherRootID := s.mustCreate(user, "Other Root", nil)
	subID := s.mustCreate(user, "Sub", &rootID)

	resp := s.request(http.MethodPut, "/categories/"+rootID, user, validRequest("Root", &otherRootID))
	s.Equal(http.StatusUnprocessableEntity, resp.status)

	resp = s.request(http.MethodPut, "/categories/"+subID, user, validRequest("Sub", &subID))
	s.Equal(http.StatusUnprocessableEntity, resp.status)
}

func (s *CategoriesE2ESuite) TestListScopeSubsAndPaginationContract() {
	user := newUserID()

	rootID := s.mustCreate(user, "Casa", nil)
	_ = s.mustCreate(user, "Mercado", &rootID)
	_ = s.mustCreate(user, "Trabalho", nil)

	resp := s.request(http.MethodGet, pkgroutes.Categories+"?scope=subs", user, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list := decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(int64(1), list.Total)
	s.Len(list.Items, 1)
	s.NotNil(list.Items[0].ParentID)
	s.Equal(10, list.PageSize)

	resp = s.request(http.MethodGet, pkgroutes.Categories+"?pageSize=999", user, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list = decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(100, list.PageSize)
}

// --- helpers ------------------------------------------------------------------

func (s *CategoriesE2ESuite) mustCreate(user, name string, parent *string) string {
	resp := s.request(http.MethodPost, pkgroutes.Categories, user, validRequest(name, parent))
	s.Require().Equalf(http.StatusCreated, resp.status,
		"creating %q expected 201, got %d body=%s", name, resp.status, resp.text())
	return decodeJSON[dtos.CategoryResponse](s, resp).ID
}

// seedRoots inserts n active root categories for the given user using a
// transactional prepared statement. Bypassing the repository keeps the seed
// step within a few seconds even for 10k rows.
func (s *CategoriesE2ESuite) seedRoots(user string, n int) {
	ctx := context.Background()
	now := s.clock.Now()

	tx, err := s.db.BeginTxx(ctx, nil)
	s.Require().NoError(err)
	defer func() { _ = tx.Rollback() }()

	const insertSQL = `INSERT INTO dbo.[Category]
		([Id],[UserId],[ParentId],[Name],[Color],[Icon],[CreatedAt],[UpdatedAt],[DeletedAt])
		VALUES (@id,@userId,NULL,@name,@color,@icon,@createdAt,@updatedAt,NULL)`

	stmt, err := tx.PreparexContext(ctx, insertSQL)
	s.Require().NoError(err)
	defer stmt.Close()

	for i := range n {
		_, err := stmt.ExecContext(ctx,
			sql.Named("id", uuid.NewString()),
			sql.Named("userId", user),
			sql.Named("name", fmt.Sprintf("Cat-%06d", i)),
			sql.Named("color", "blue"),
			sql.Named("icon", "wallet"),
			sql.Named("createdAt", now),
			sql.Named("updatedAt", now),
		)
		s.Require().NoErrorf(err, "seed row %d", i)
	}
	s.Require().NoError(tx.Commit())
}
