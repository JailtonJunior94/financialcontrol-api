package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name               string
		page, size         int
		wantPage, wantSize int
		wantEnabled        bool
		wantOffset         int
	}{
		{"zero desabilita", 0, 0, 0, 0, false, 0},
		{"normaliza page<=0", -1, 10, 1, 10, true, 0},
		{"normaliza size<=0", 2, -5, 2, 10, true, 10},
		{"clamp size", 1, 1000, 1, 100, true, 0},
		{"valores válidos", 3, 25, 3, 25, true, 50},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := vos.NewPagination(tc.page, tc.size)
			assert.Equal(t, tc.wantPage, p.Page)
			assert.Equal(t, tc.wantSize, p.Size)
			assert.Equal(t, tc.wantEnabled, p.Enabled())
			assert.Equal(t, tc.wantOffset, p.Offset())
		})
	}
}
