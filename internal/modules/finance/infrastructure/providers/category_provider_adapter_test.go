package providers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	categoriesdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	categoriesentities "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
	categoriesmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports/mocks"
	categoriesvos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	financedomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	financevos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/providers"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	testCategoryIDStr = "44444444-4444-4444-8444-444444444444"
)

func makeTestCategory(t *testing.T, active bool) *categoriesentities.Category {
	t.Helper()

	id, err := categoriesvos.ParseCategoryID(testCategoryIDStr)
	require.NoError(t, err)

	name, err := categoriesvos.NewCategoryName("Alimentação")
	require.NoError(t, err)

	now := time.Now().UTC()
	return categoriesentities.RehydrateCategory(id, name, 1, now, now, active)
}

func TestCategoryProviderAdapter_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbErr := errors.New("db error")

	userID, err := identityvo.ParseUserID(testUserIDStr)
	require.NoError(t, err)

	categoryID, err := categoriesvos.ParseCategoryID(testCategoryIDStr)
	require.NoError(t, err)

	financeCategoryID, err := financevos.ParseCategoryID(testCategoryIDStr)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupMock   func(t *testing.T, repo *categoriesmocks.CategoryRepository)
		wantActive  bool
		wantName    string
		wantErr     error
		wantErrWrap bool
	}{
		{
			name: "happy path — active category",
			setupMock: func(t *testing.T, repo *categoriesmocks.CategoryRepository) {
				cat := makeTestCategory(t, true)
				repo.On("GetByID", ctx, userID, categoryID).Return(cat, nil)
			},
			wantActive: true,
			wantName:   "Alimentação",
		},
		{
			name: "inactive category — Active=false, nil error (G2.a)",
			setupMock: func(t *testing.T, repo *categoriesmocks.CategoryRepository) {
				cat := makeTestCategory(t, false)
				repo.On("GetByID", ctx, userID, categoryID).Return(cat, nil)
			},
			wantActive: false,
			wantName:   "Alimentação",
		},
		{
			name: "category not found or belongs to another user — ErrCategoryNotFound (G2.a)",
			setupMock: func(t *testing.T, repo *categoriesmocks.CategoryRepository) {
				repo.On("GetByID", ctx, userID, categoryID).
					Return((*categoriesentities.Category)(nil), categoriesdomain.ErrCategoryNotFound)
			},
			wantErr: financedomain.ErrCategoryNotFound,
		},
		{
			name: "DB error — propagated wrapped",
			setupMock: func(t *testing.T, repo *categoriesmocks.CategoryRepository) {
				repo.On("GetByID", ctx, userID, categoryID).
					Return((*categoriesentities.Category)(nil), dbErr)
			},
			wantErrWrap: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := categoriesmocks.NewCategoryRepository(t)
			tc.setupMock(t, repo)

			adapter := providers.NewCategoryProviderAdapter(repo)
			got, err := adapter.GetByID(ctx, userID, financeCategoryID)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, projections.CategoryView{}, got)
				return
			}
			if tc.wantErrWrap {
				require.Error(t, err)
				assert.ErrorContains(t, err, "category provider")
				assert.Equal(t, projections.CategoryView{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, financeCategoryID, got.ID)
			assert.Equal(t, userID, got.UserID)
			assert.Equal(t, tc.wantName, got.Name)
			assert.Nil(t, got.ParentID)
			assert.Equal(t, tc.wantActive, got.Active)
		})
	}
}
