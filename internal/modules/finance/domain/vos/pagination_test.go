package vos_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
)

func TestNewPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		rawPage      int
		rawPageSize  int
		wantPage     int
		wantPageSize int
		wantErr      error
	}{
		{name: "defaults when both zero", rawPage: 0, rawPageSize: 0, wantPage: 1, wantPageSize: 20},
		{name: "page 0 defaults to 1", rawPage: 0, rawPageSize: 10, wantPage: 1, wantPageSize: 10},
		{name: "page -1 defaults to 1", rawPage: -1, rawPageSize: 10, wantPage: 1, wantPageSize: 10},
		{name: "page_size 0 defaults to 20", rawPage: 1, rawPageSize: 0, wantPage: 1, wantPageSize: 20},
		{name: "page_size -1 defaults to 20", rawPage: 1, rawPageSize: -1, wantPage: 1, wantPageSize: 20},
		{name: "valid mid-range", rawPage: 3, rawPageSize: 50, wantPage: 3, wantPageSize: 50},
		{name: "max page_size 100", rawPage: 1, rawPageSize: 100, wantPage: 1, wantPageSize: 100},
		{name: "page_size 101 is invalid", rawPage: 1, rawPageSize: 101, wantErr: domain.ErrInvalidPaginationLimits},
		{name: "page_size 200 is invalid", rawPage: 1, rawPageSize: 200, wantErr: domain.ErrInvalidPaginationLimits},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pg, err := vos.NewPagination(tc.rawPage, tc.rawPageSize)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPage, pg.Page())
			assert.Equal(t, tc.wantPageSize, pg.PageSize())
		})
	}
}

func TestPaginationOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		page     int
		pageSize int
		want     int
	}{
		{page: 1, pageSize: 20, want: 0},
		{page: 2, pageSize: 20, want: 20},
		{page: 3, pageSize: 10, want: 20},
		{page: 5, pageSize: 100, want: 400},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			pg, err := vos.NewPagination(tc.page, tc.pageSize)
			require.NoError(t, err)
			assert.Equal(t, tc.want, pg.Offset())
		})
	}
}
