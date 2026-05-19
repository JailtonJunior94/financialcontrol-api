package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func TestCategoryUniquenessService_EnsureUnique(t *testing.T) {
	userID, err := identityvo.ParseUserID("11111111-1111-4111-8111-111111111111")
	require.NoError(t, err)
	name, err := vos.NewCategoryName("comida")
	require.NoError(t, err)
	parentID := vos.NewCategoryID()
	excludeID := vos.NewCategoryID()
	boom := errors.New("db error")

	tests := []struct {
		name      string
		parent    *vos.CategoryID
		exclude   *vos.CategoryID
		repoExist bool
		repoErr   error
		wantErr   error
	}{
		{
			name:      "root scope free",
			parent:    nil,
			exclude:   nil,
			repoExist: false,
		},
		{
			name:      "root scope conflict",
			parent:    nil,
			exclude:   nil,
			repoExist: true,
			wantErr:   domain.ErrCategoryNameAlreadyExists,
		},
		{
			name:      "sub scope free",
			parent:    &parentID,
			exclude:   nil,
			repoExist: false,
		},
		{
			name:      "sub scope conflict",
			parent:    &parentID,
			exclude:   nil,
			repoExist: true,
			wantErr:   domain.ErrCategoryNameAlreadyExists,
		},
		{
			name:      "update ignores own row via excludeID",
			parent:    nil,
			exclude:   &excludeID,
			repoExist: false,
		},
		{
			name:    "propagates repository error",
			parent:  nil,
			exclude: nil,
			repoErr: boom,
			wantErr: boom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewCategoryRepository(t)
			repo.EXPECT().
				ExistsByName(mock.Anything, userID, name, tc.parent, tc.exclude).
				Return(tc.repoExist, tc.repoErr)

			svc := services.NewCategoryUniquenessService(repo)
			err := svc.EnsureUnique(context.Background(), userID, name, tc.parent, tc.exclude)
			if tc.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
