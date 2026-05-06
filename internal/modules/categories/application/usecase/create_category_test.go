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
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
)

func ptr[T any](v T) *T { return &v }

func TestCreateCategory_Execute(t *testing.T) {
	userID := mustUserID(t)
	clock := newFixedClock()

	tests := []struct {
		name   string
		req    dtos.CategoryRequest
		setup  func(repo *mocks.CategoryRepository)
		assert func(t *testing.T, resp dtos.CategoryResponse, err error)
	}{
		{
			name: "golden path root",
			req:  dtos.CategoryRequest{Name: "Comida", Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, resp dtos.CategoryResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Comida", resp.Name)
				assert.Equal(t, "red", resp.Color)
				assert.Equal(t, "ic-food", resp.Icon)
				assert.Nil(t, resp.ParentID)
			},
		},
		{
			name: "golden path subcategory",
			req: dtos.CategoryRequest{
				Name: "Mercado", Color: "blue", Icon: "ic-cart",
				ParentID: ptr("22222222-2222-4222-8222-222222222222"),
			},
			setup: func(repo *mocks.CategoryRepository) {
				parentID, _ := vos.ParseCategoryID("22222222-2222-4222-8222-222222222222")
				parent, _ := entities.NewCategory(userID, nil, "Comida", "red", "ic-food", clock)
				_ = parent
				// rebuild parent with the explicit id
				name, _ := vos.NewCategoryName("Comida")
				color, _ := vos.NewCategoryColor("red")
				icon, _ := vos.NewCategoryIcon("ic-food")
				rehydrated := entities.RehydrateCategory(parentID, userID, nil, name, color, icon, clock.Now(), clock.Now(), nil)
				repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, parentID).Return(rehydrated, nil).Once()
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, &parentID, (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, resp dtos.CategoryResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp.ParentID)
				assert.Equal(t, "22222222-2222-4222-8222-222222222222", *resp.ParentID)
			},
		},
		{
			name: "invalid name",
			req:  dtos.CategoryRequest{Name: "", Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryName)
			},
		},
		{
			name: "invalid parent id format",
			req: dtos.CategoryRequest{
				Name: "Mercado", Color: "red", Icon: "ic-food",
				ParentID: ptr("not-a-uuid"),
			},
			setup: func(repo *mocks.CategoryRepository) {
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryID)
			},
		},
		{
			name: "parent not found (cross user maps to parent not found)",
			req: dtos.CategoryRequest{
				Name: "Mercado", Color: "red", Icon: "ic-food",
				ParentID: ptr("22222222-2222-4222-8222-222222222222"),
			},
			setup: func(repo *mocks.CategoryRepository) {
				parentID, _ := vos.ParseCategoryID("22222222-2222-4222-8222-222222222222")
				repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, parentID).Return(nil, domain.ErrCategoryNotFound).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrParentNotFound)
			},
		},
		{
			name: "parent inactive",
			req: dtos.CategoryRequest{
				Name: "Mercado", Color: "red", Icon: "ic-food",
				ParentID: ptr("22222222-2222-4222-8222-222222222222"),
			},
			setup: func(repo *mocks.CategoryRepository) {
				parentID, _ := vos.ParseCategoryID("22222222-2222-4222-8222-222222222222")
				name, _ := vos.NewCategoryName("Comida")
				color, _ := vos.NewCategoryColor("red")
				icon, _ := vos.NewCategoryIcon("ic-food")
				del := clock.Now()
				rehydrated := entities.RehydrateCategory(parentID, userID, nil, name, color, icon, clock.Now(), clock.Now(), &del)
				repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, parentID).Return(rehydrated, nil).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrParentInactive)
			},
		},
		{
			name: "parent is itself a subcategory -> depth exceeded",
			req: dtos.CategoryRequest{
				Name: "Detail", Color: "red", Icon: "ic-food",
				ParentID: ptr("22222222-2222-4222-8222-222222222222"),
			},
			setup: func(repo *mocks.CategoryRepository) {
				parentID, _ := vos.ParseCategoryID("22222222-2222-4222-8222-222222222222")
				grand, _ := vos.ParseCategoryID("33333333-3333-4333-8333-333333333333")
				name, _ := vos.NewCategoryName("Mercado")
				color, _ := vos.NewCategoryColor("red")
				icon, _ := vos.NewCategoryIcon("ic-food")
				rehydrated := entities.RehydrateCategory(parentID, userID, &grand, name, color, icon, clock.Now(), clock.Now(), nil)
				repo.EXPECT().GetByIDIncludingDeleted(mock.Anything, userID, parentID).Return(rehydrated, nil).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrSubcategoryDepthExceeded)
			},
		},
		{
			name: "duplicate name",
			req:  dtos.CategoryRequest{Name: "Comida", Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(true, nil).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrCategoryNameAlreadyExists)
			},
		},
		{
			name: "uniqueness check repo error",
			req:  dtos.CategoryRequest{Name: "Comida", Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, errors.New("db down")).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "repo add error",
			req:  dtos.CategoryRequest{Name: "Comida", Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(errors.New("boom")).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid color",
			req:  dtos.CategoryRequest{Name: "Comida", Color: "fluor", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryColor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewCategoryRepository(t)
			tt.setup(repo)
			uniq := services.NewCategoryUniquenessService(repo)
			sut := usecase.NewCreateCategory(repo, uniq, clock)
			resp, err := sut.Execute(context.Background(), userID, tt.req)
			tt.assert(t, resp, err)
		})
	}
}
