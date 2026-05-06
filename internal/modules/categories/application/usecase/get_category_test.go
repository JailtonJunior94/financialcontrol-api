package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func TestGetCategory_Execute(t *testing.T) {
	userID := mustUserID(t)
	idStr := "44444444-4444-4444-8444-444444444444"
	id, _ := vos.ParseCategoryID(idStr)

	t.Run("invalid id", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		_, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, "bad")
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
	})

	t.Run("root", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		root := makeCategory(t, id, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(root, nil).Once()
		resp, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, idStr)
		require.NoError(t, err)
		assert.Nil(t, resp.Parent)
	})

	t.Run("subcategory hydrates parent", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		parentID, _ := vos.ParseCategoryID("55555555-5555-4555-8555-555555555555")
		sub := makeCategory(t, id, &parentID)
		parent := makeCategory(t, parentID, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(sub, nil).Once()
		repo.EXPECT().GetByID(mock.Anything, userID, parentID).Return(parent, nil).Once()
		resp, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, idStr)
		require.NoError(t, err)
		require.NotNil(t, resp.Parent)
		assert.Equal(t, parentID.String(), resp.Parent.ID)
	})

	t.Run("not found cross user", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, domain.ErrCategoryNotFound).Once()
		_, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, idStr)
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, errors.New("boom")).Once()
		_, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, idStr)
		assert.Error(t, err)
	})
}
