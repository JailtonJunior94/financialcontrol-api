package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports/mocks"
)

func TestListCategories_Execute(t *testing.T) {
	userID := mustUserID(t)

	t.Run("default pagination", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f ports.ListFilter) bool {
			return f.Pagination.Page == 1 && f.Pagination.Size == 10 && f.NameLike == ""
		})).Return([]entities.Category{}, int64(0), nil).Once()

		resp, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
	})

	t.Run("roots scope is accepted as all", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.Anything).Return([]entities.Category{}, int64(0), nil).Once()

		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "roots"})
		require.NoError(t, err)
	})

	t.Run("subs scope is unsupported", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "subs"})
		assert.ErrorIs(t, err, domain.ErrCategoryHierarchyUnsupported)
	})

	t.Run("parentId is unsupported", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{
			ParentID: "55555555-5555-4555-8555-555555555555",
		})
		assert.ErrorIs(t, err, domain.ErrCategoryHierarchyUnsupported)
	})

	t.Run("custom pagination and filter", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f ports.ListFilter) bool {
			return f.Pagination.Page == 3 && f.Pagination.Size == 10 && f.NameLike == "co"
		})).Return([]entities.Category{}, int64(42), nil).Once()

		resp, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Page: 3, PageSize: 10, Name: "co"})
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp.Total)
		assert.Equal(t, 3, resp.Page)
	})

	t.Run("page size capped at one hundred", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f ports.ListFilter) bool {
			return f.Pagination.Page == 1 && f.Pagination.Size == 100
		})).Return([]entities.Category{}, int64(0), nil).Once()

		resp, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{PageSize: 999})
		require.NoError(t, err)
		assert.Equal(t, 100, resp.PageSize)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.Anything).Return(nil, int64(0), errors.New("boom")).Once()
		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{})
		assert.Error(t, err)
	})
}
