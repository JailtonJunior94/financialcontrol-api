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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func TestListCategories_Execute(t *testing.T) {
	userID := mustUserID(t)

	t.Run("default pagination + scope=all", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return f.Pagination.Page == 1 && f.Pagination.Size == 10 && !f.OnlyRoots && !f.OnlySubs && f.ParentID == nil
		})).Return([]entities.Category{}, int64(0), nil).Once()

		resp, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
	})

	t.Run("scope roots", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return f.OnlyRoots && f.ParentID == nil
		})).Return([]entities.Category{}, int64(0), nil).Once()

		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "roots"})
		require.NoError(t, err)
	})

	t.Run("scope subs with parentId", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		parentIDStr := "55555555-5555-4555-8555-555555555555"
		parentID, _ := vos.ParseCategoryID(parentIDStr)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return !f.OnlyRoots && f.OnlySubs && f.ParentID != nil && *f.ParentID == parentID
		})).Return([]entities.Category{}, int64(0), nil).Once()

		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "subs", ParentID: parentIDStr})
		require.NoError(t, err)
	})

	t.Run("scope subs without parentId", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return !f.OnlyRoots && f.OnlySubs && f.ParentID == nil
		})).Return([]entities.Category{}, int64(0), nil).Once()

		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "subs"})
		require.NoError(t, err)
	})

	t.Run("parentId overrides roots scope", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		parentIDStr := "55555555-5555-4555-8555-555555555555"
		parentID, _ := vos.ParseCategoryID(parentIDStr)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return !f.OnlyRoots && f.OnlySubs && f.ParentID != nil && *f.ParentID == parentID
		})).Return([]entities.Category{}, int64(0), nil).Once()

		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Scope: "roots", ParentID: parentIDStr})
		require.NoError(t, err)
	})

	t.Run("invalid parentId", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		_, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{ParentID: "bad"})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
	})

	t.Run("custom pagination", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
			return f.Pagination.Page == 3 && f.Pagination.Size == 10 && f.NameLike == "co"
		})).Return([]entities.Category{}, int64(42), nil).Once()

		resp, err := usecase.NewListCategories(repo).Execute(context.Background(), userID, dtos.ListCategoriesQuery{Page: 3, PageSize: 10, Name: "co"})
		require.NoError(t, err)
		assert.Equal(t, int64(42), resp.Total)
		assert.Equal(t, 3, resp.Page)
	})

	t.Run("page size is capped at one hundred", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(f interfaces.ListFilter) bool {
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
