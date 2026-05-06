package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func mustUserID(t *testing.T) identityvo.UserID {
	t.Helper()
	id, err := identityvo.ParseUserID("11111111-1111-4111-8111-111111111111")
	require.NoError(t, err)
	return id
}

func newCategory(t *testing.T, parent *vos.CategoryID, deleted bool) *entities.Category {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	id := vos.NewCategoryID()
	name, err := vos.NewCategoryName("comida")
	require.NoError(t, err)
	color, err := vos.NewCategoryColor("red")
	require.NoError(t, err)
	icon, err := vos.NewCategoryIcon("ic-food")
	require.NoError(t, err)

	var deletedAt *time.Time
	if deleted {
		d := now
		deletedAt = &d
	}
	return entities.RehydrateCategory(id, mustUserID(t), parent, name, color, icon, now, now, deletedAt)
}

func TestCategoryDeletionService_Delete(t *testing.T) {
	userID := mustUserID(t)
	fixed := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	clock := fixedClock{now: fixed}

	t.Run("root cascade", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		root := newCategory(t, nil, false)
		repo.EXPECT().GetByID(mock.Anything, userID, root.ID()).Return(root, nil)
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, root.ID(), fixed).Return(nil)

		svc := services.NewCategoryDeletionService(repo, clock)
		require.NoError(t, svc.Delete(context.Background(), userID, root.ID()))
	})

	t.Run("standalone subcategory", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		parentID := vos.NewCategoryID()
		sub := newCategory(t, &parentID, false)
		repo.EXPECT().GetByID(mock.Anything, userID, sub.ID()).Return(sub, nil)
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, sub.ID(), fixed).Return(nil)

		svc := services.NewCategoryDeletionService(repo, clock)
		require.NoError(t, svc.Delete(context.Background(), userID, sub.ID()))
	})

	t.Run("propagates domain not found", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		id := vos.NewCategoryID()
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, domain.ErrCategoryNotFound)

		svc := services.NewCategoryDeletionService(repo, clock)
		err := svc.Delete(context.Background(), userID, id)
		assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
	})

	t.Run("propagates repository error on lookup", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		boom := errors.New("db down")
		id := vos.NewCategoryID()
		repo.EXPECT().GetByID(mock.Anything, userID, id).Return(nil, boom)

		svc := services.NewCategoryDeletionService(repo, clock)
		err := svc.Delete(context.Background(), userID, id)
		assert.ErrorIs(t, err, boom)
	})

	t.Run("propagates repository error on cascade", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		root := newCategory(t, nil, false)
		boom := errors.New("tx aborted")
		repo.EXPECT().GetByID(mock.Anything, userID, root.ID()).Return(root, nil)
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, root.ID(), fixed).Return(boom)

		svc := services.NewCategoryDeletionService(repo, clock)
		err := svc.Delete(context.Background(), userID, root.ID())
		assert.ErrorIs(t, err, boom)
	})
}
