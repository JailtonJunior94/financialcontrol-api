//go:build integration

package categories_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

const ddlCategoryTable = `
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

type stubParser struct{}

func (stubParser) Parse(_ context.Context, token string) (pkgjwt.Identity, error) {
	if token == "" {
		return pkgjwt.Identity{}, fmt.Errorf("empty token")
	}
	return pkgjwt.Identity{UserID: token, Email: token + "@example.com"}, nil
}

type CategoriesE2ESuite struct {
	suite.Suite

	mgr   devkitmgr.Manager
	db    devkitdb.DBTX
	app   *fiber.App
	clock fixedClock
}

func TestCategoriesE2E(t *testing.T) {
	suite.Run(t, new(CategoriesE2ESuite))
}

func (s *CategoriesE2ESuite) SetupSuite() {
	mgr, _, err := dbmssql.GetSharedTestManager()
	s.Require().NoError(err)
	s.mgr = mgr
	s.db = mgr.DBTX(context.Background())

	_, err = s.db.ExecContext(context.Background(), ddlCategoryTable)
	s.Require().NoError(err)

	s.app = s.buildApp()
}

func (s *CategoriesE2ESuite) SetupTest() {
	s.clock = fixedClock{now: time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC)}
	s.truncate()
	s.app = s.buildApp()
}

func (s *CategoriesE2ESuite) truncate() {
	_, err := s.db.ExecContext(context.Background(), `DELETE FROM dbo.[Category]`)
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

func validRequest(name string, sequence int) dtos.CategoryRequest {
	return dtos.CategoryRequest{Name: name, Sequence: sequence, Color: "blue", Icon: "wallet"}
}

func (s *CategoriesE2ESuite) TestFullLifecycle() {
	user := newUserID()

	resp := s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Mercado", 3))
	s.Require().Equal(http.StatusCreated, resp.status)
	category := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Require().NotEmpty(category.ID)
	s.Equal(3, category.Sequence)

	resp = s.request(http.MethodGet, pkgroutes.Categories+"?pageSize=10", user, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list := decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(int64(1), list.Total)
	s.Len(list.Items, 1)

	resp = s.request(http.MethodPut, "/categories/"+category.ID, user, validRequest("Mercado Atualizado", 9))
	s.Require().Equal(http.StatusOK, resp.status)
	updated := decodeJSON[dtos.CategoryResponse](s, resp)
	s.Equal("Mercado Atualizado", updated.Name)
	s.Equal(9, updated.Sequence)

	resp = s.request(http.MethodDelete, "/categories/"+category.ID, user, nil)
	s.Require().Equal(http.StatusNoContent, resp.status)

	resp = s.request(http.MethodGet, "/categories/"+category.ID, user, nil)
	s.Equal(http.StatusNotFound, resp.status)

	resp = s.request(http.MethodGet, pkgroutes.Categories, user, nil)
	s.Require().Equal(http.StatusOK, resp.status)
	list = decodeJSON[dtos.PaginatedResponse[dtos.CategoryResponse]](s, resp)
	s.Equal(int64(0), list.Total)
}

func (s *CategoriesE2ESuite) TestUnsupportedHierarchyAndNameReuseAfterDelete() {
	user := newUserID()

	resp := s.request(http.MethodPost, pkgroutes.Categories, user, dtos.CategoryRequest{
		Name: "Casa", Sequence: 1, ParentID: ptr(uuid.NewString()),
	})
	s.Equal(http.StatusUnprocessableEntity, resp.status)

	resp = s.request(http.MethodGet, pkgroutes.Categories+"?scope=subs", user, nil)
	s.Equal(http.StatusUnprocessableEntity, resp.status)

	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", 1))
	s.Require().Equal(http.StatusCreated, resp.status)
	first := decodeJSON[dtos.CategoryResponse](s, resp)

	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", 2))
	s.Equal(http.StatusConflict, resp.status)

	resp = s.request(http.MethodDelete, "/categories/"+first.ID, user, nil)
	s.Require().Equal(http.StatusNoContent, resp.status)

	resp = s.request(http.MethodPost, pkgroutes.Categories, user, validRequest("Casa", 2))
	s.Equal(http.StatusCreated, resp.status)
}

func ptr[T any](v T) *T { return &v }
