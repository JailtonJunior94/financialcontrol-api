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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports/mocks"
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

func newCategory(t *testing.T, active bool) *entities.Category {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	id := vos.NewCategoryID()
	name, err := vos.NewCategoryName("comida")
	require.NoError(t, err)
	return entities.RehydrateCategory(id, name, 1, now, now, active)
}

func TestCategoryDeletionService_Delete(t *testing.T) {
	userID := mustUserID(t)
	fixed := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	clock := fixedClock{now: fixed}

	t.Run("deactivates active category", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		category := newCategory(t, true)
		repo.EXPECT().GetByID(mock.Anything, userID, category.ID()).Return(category, nil)
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, category.ID(), fixed).Return(nil)

		svc := services.NewCategoryDeletionService(repo, clock)
		require.NoError(t, svc.Delete(context.Background(), userID, category.ID()))
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

	t.Run("propagates repository error on deactivate", func(t *testing.T) {
		repo := mocks.NewCategoryRepository(t)
		category := newCategory(t, true)
		boom := errors.New("tx aborted")
		repo.EXPECT().GetByID(mock.Anything, userID, category.ID()).Return(category, nil)
		repo.EXPECT().SoftDeleteCascade(mock.Anything, userID, category.ID(), fixed).Return(boom)

		svc := services.NewCategoryDeletionService(repo, clock)
		err := svc.Delete(context.Background(), userID, category.ID())
		assert.ErrorIs(t, err, boom)
	})
}
