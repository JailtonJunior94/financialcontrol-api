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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports/mocks"
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

	t.Run("found", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		category := makeCategory(t, id, 7, true)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(category, nil).Once()
		resp, err := usecase.NewGetCategory(repo).Execute(context.Background(), userID, idStr)
		require.NoError(t, err)
		assert.Equal(t, 7, resp.Sequence)
		assert.True(t, resp.Active)
	})

	t.Run("not found", func(t *testing.T) {
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
