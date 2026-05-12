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
			name: "golden path",
			req:  dtos.CategoryRequest{Name: "Comida", Sequence: 3, Color: "red", Icon: "ic-food"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, resp dtos.CategoryResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, "Comida", resp.Name)
				assert.Equal(t, 3, resp.Sequence)
				assert.True(t, resp.Active)
			},
		},
		{
			name: "hierarchy unsupported",
			req: dtos.CategoryRequest{
				Name: "Mercado", Sequence: 1,
				ParentID: ptr("22222222-2222-4222-8222-222222222222"),
			},
			setup: func(repo *mocks.CategoryRepository) {},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrCategoryHierarchyUnsupported)
			},
		},
		{
			name:  "invalid name",
			req:   dtos.CategoryRequest{Name: "", Sequence: 1},
			setup: func(repo *mocks.CategoryRepository) {},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryName)
			},
		},
		{
			name: "duplicate name",
			req:  dtos.CategoryRequest{Name: "Comida", Sequence: 1},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(true, nil).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.ErrorIs(t, err, domain.ErrCategoryNameAlreadyExists)
			},
		},
		{
			name: "legacy styling fields are ignored",
			req:  dtos.CategoryRequest{Name: "Comida", Sequence: 2, Color: "fluor", Icon: "!!!"},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(nil).Once()
			},
			assert: func(t *testing.T, resp dtos.CategoryResponse, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2, resp.Sequence)
			},
		},
		{
			name: "uniqueness check repo error",
			req:  dtos.CategoryRequest{Name: "Comida", Sequence: 1},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, errors.New("db down")).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "repo add error",
			req:  dtos.CategoryRequest{Name: "Comida", Sequence: 1},
			setup: func(repo *mocks.CategoryRepository) {
				repo.EXPECT().ExistsByName(mock.Anything, userID, mock.Anything, (*vos.CategoryID)(nil), (*vos.CategoryID)(nil)).Return(false, nil).Once()
				repo.EXPECT().Add(mock.Anything, mock.Anything).Return(errors.New("boom")).Once()
			},
			assert: func(t *testing.T, _ dtos.CategoryResponse, err error) {
				assert.Error(t, err)
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
