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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func makeCategory(t *testing.T, id vos.CategoryID, parentID *vos.CategoryID) *entities.Category {
	t.Helper()
	name, err := vos.NewCategoryName("Comida")
	require.NoError(t, err)
	color, err := vos.NewCategoryColor("red")
	require.NoError(t, err)
	icon, err := vos.NewCategoryIcon("ic-food")
	require.NoError(t, err)
	now := newFixedClock().Now()
	return entities.RehydrateCategory(id, mustUserID(t), parentID, name, color, icon, now, now, nil)
}

func TestUpdateCategory_Execute(t *testing.T) {
	userID := mustUserID(t)
	clock := newFixedClock()
	idStr := "44444444-4444-4444-8444-444444444444"
	id, _ := vos.ParseCategoryID(idStr)

	t.Run("invalid id", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, "bad", dtos.CategoryRequest{Name: "x", Color: "red", Icon: "ic-food"})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
	})

	t.Run("not found is propagated (cross user)", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, domain.ErrCategoryNotFound).Once()
		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "x", Color: "red", Icon: "ic-food"})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("rename + change appearance", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(false, nil).Once()
		repo.EXPECT().Update(mock.Anything, current).Return(nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		resp, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Mercado", Color: "blue", Icon: "ic-cart"})
		require.NoError(t, err)
		assert.Equal(t, "Mercado", resp.Name)
		assert.Equal(t, "blue", resp.Color)
		assert.Equal(t, "ic-cart", resp.Icon)
	})

	t.Run("reparent to inactive parent", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		oldParent := vos.NewCategoryID()
		current := makeCategory(t, id, &oldParent)
		newParentIDStr := "55555555-5555-4555-8555-555555555555"
		newParentID, _ := vos.ParseCategoryID(newParentIDStr)

		// inactive parent
		name, _ := vos.NewCategoryName("Pai")
		color, _ := vos.NewCategoryColor("red")
		icon, _ := vos.NewCategoryIcon("ic-food")
		del := clock.Now()
		inactive := entities.RehydrateCategory(newParentID, userID, nil, name, color, icon, clock.Now(), clock.Now(), &del)

		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, newParentID).Return(inactive, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{
			Name: "x", Color: "red", Icon: "ic-food",
			ParentID: ptr(newParentIDStr),
		})
		assert.ErrorIs(t, err, domain.ErrParentInactive)
	})

	t.Run("reparent to non-existent parent", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		oldParent := vos.NewCategoryID()
		current := makeCategory(t, id, &oldParent)
		newParentIDStr := "55555555-5555-4555-8555-555555555555"
		newParentID, _ := vos.ParseCategoryID(newParentIDStr)

		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, newParentID).Return(nil, domain.ErrCategoryNotFound).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{
			Name: "x", Color: "red", Icon: "ic-food",
			ParentID: ptr(newParentIDStr),
		})
		assert.ErrorIs(t, err, domain.ErrParentNotFound)
	})

	t.Run("name conflict in scope", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(true, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Outro", Color: "red", Icon: "ic-food"})
		assert.ErrorIs(t, err, domain.ErrCategoryNameAlreadyExists)
	})

	t.Run("invalid color", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "x", Color: "fluor", Icon: "ic-food"})
		assert.ErrorIs(t, err, domain.ErrInvalidCategoryColor)
	})

	t.Run("root cannot become subcategory", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, nil)
		newParentIDStr := "55555555-5555-4555-8555-555555555555"

		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{
			Name: "Mercado", Color: "red", Icon: "ic-food",
			ParentID: ptr(newParentIDStr),
		})
		assert.ErrorIs(t, err, domain.ErrSubcategoryDepthExceeded)
	})

	t.Run("category cannot be reparented to itself", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		parentID := vos.NewCategoryID()
		current := makeCategory(t, id, &parentID)

		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{
			Name: "Mercado", Color: "red", Icon: "ic-food",
			ParentID: ptr(idStr),
		})
		assert.ErrorIs(t, err, domain.ErrSubcategoryDepthExceeded)
	})

	t.Run("update returns not found when row disappears before persistence", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		current := makeCategory(t, id, nil)
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(current, nil).Once()
		repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), &id).Return(false, nil).Once()
		repo.EXPECT().Update(mock.Anything, current).Return(domain.ErrCategoryNotFound).Once()

		uniq := services.NewCategoryUniquenessService(repo)
		sut := usecase.NewUpdateCategory(repo, uniq, clock)
		_, err := sut.Execute(context.Background(), userID, idStr, dtos.CategoryRequest{Name: "Mercado", Color: "blue", Icon: "ic-cart"})
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})
}
