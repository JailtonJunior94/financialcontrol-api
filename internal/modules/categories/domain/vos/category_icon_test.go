package vos_test

import (
	"strings"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategoryIcon(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"válido simples", "food", "food", false},
		{"válido com hífen e dígito", "car-1", "car-1", false},
		{"64 chars", strings.Repeat("a", 64), strings.Repeat("a", 64), false},
		{"65 chars", strings.Repeat("a", 65), "", true},
		{"vazio", "", "", true},
		{"upper case rejeitado", "Food", "", true},
		{"underscore rejeitado", "fast_food", "", true},
		{"espaço interno rejeitado", "fast food", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewCategoryIcon(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryIcon)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}
