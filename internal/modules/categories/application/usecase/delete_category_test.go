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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func TestDeleteCategory_Execute(t *testing.T) {
	userID := mustUserID(t)
	clock := newFixedClock()
	idStr := "44444444-4444-4444-8444-444444444444"
	id, _ := vos.ParseCategoryID(idStr)

	t.Run("invalid id", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		svc := services.NewCategoryDeletionService(repo, clock)
		err := usecase.NewDeleteCategory(svc).Execute(context.Background(), userID, "bad")
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
	})

	t.Run("deactivates existing category", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		category := makeCategory(t, id, 1, true)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(category, nil).Once()
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, id, clock.Now()).Return(nil).Once()
		svc := services.NewCategoryDeletionService(repo, clock)
		require.NoError(t, usecase.NewDeleteCategory(svc).Execute(context.Background(), userID, idStr))
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, domain.ErrCategoryNotFound).Once()
		svc := services.NewCategoryDeletionService(repo, clock)
		err := usecase.NewDeleteCategory(svc).Execute(context.Background(), userID, idStr)
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("repo error wrapped", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, errors.New("boom")).Once()
		svc := services.NewCategoryDeletionService(repo, clock)
		err := usecase.NewDeleteCategory(svc).Execute(context.Background(), userID, idStr)
		assert.Error(t, err)
	})
}
