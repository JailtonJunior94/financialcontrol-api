package dtos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
)

func TestNewPagination(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		page       int
		size       int
		wantPage   int
		wantSize   int
		wantOffset int
	}{
		{name: "zero value preserva listagem sem paginacao", page: 0, size: 0, wantPage: 0, wantSize: 0, wantOffset: 0},
		{name: "defaults quando negativos", page: -3, size: -10, wantPage: 1, wantSize: 50, wantOffset: 0},
		{name: "valores validos", page: 2, size: 25, wantPage: 2, wantSize: 25, wantOffset: 25},
		{name: "size acima do max e clampado", page: 1, size: 1000, wantPage: 1, wantSize: 200, wantOffset: 0},
		{name: "size no limite max", page: 3, size: 200, wantPage: 3, wantSize: 200, wantOffset: 400},
		{name: "page positiva, size zero usa default", page: 5, size: 0, wantPage: 5, wantSize: 50, wantOffset: 200},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := dtos.NewPagination(tc.page, tc.size)
			assert.Equal(t, tc.wantPage, p.Page)
			assert.Equal(t, tc.wantSize, p.Size)
			assert.Equal(t, tc.wantOffset, p.Offset())
		})
	}
}
