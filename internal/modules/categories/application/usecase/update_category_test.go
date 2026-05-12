package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func TestUpdateCategory_Execute(t *testing.T) {
	userID := mustUserID(t)
	clock := newFixedClock()
	idStr := "44444444-4444-4444-8444-444444444444"
	id, _ := vos.ParseCategoryID(idStr)

	t.Run("invalid id", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, "bad", dtos.CategoryRequest{Name: "x", Sequence: 1})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
	})

	t.Run("not found", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, domain.ErrCategoryNotFound).Once()
		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "x", Sequence: 1})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("update name and sequence", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, 1, true)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(false, nil).Once()
		repo.EXPECT().Update(mock.Anything, current).Return(nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		resp, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Mercado", Sequence: 8, Color: "blue"})
		require.NoError(t, err)
		assert.Equal(t, "Mercado", resp.Name)
		assert.Equal(t, 8, resp.Sequence)
	})

	t.Run("hierarchy unsupported", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{
			Name: "Mercado", Sequence: 1, ParentID: ptr("55555555-5555-4555-8555-555555555555"),
		})
		assert.ErrorIs(t, err, domain.ErrCategoryHierarchyUnsupported)
	})

	t.Run("name conflict", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, 1, true)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(true, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Outro", Sequence: 1})
		assert.ErrorIs(t, err, domain.ErrCategoryNameAlreadyExists)
	})

	t.Run("update returns not found when row disappears", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, 1, true)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(false, nil).Once()
		repo.EXPECT().Update(mock.Anything, current).Return(domain.ErrCategoryNotFound).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Mercado", Sequence: 2})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})
}
