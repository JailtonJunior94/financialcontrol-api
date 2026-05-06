package vos_test

import (
	"strings"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategoryName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"nome válido", "Alimentação", "Alimentação", false},
		{"trim espaços", "  Lazer  ", "Lazer", false},
		{"100 chars válido", strings.Repeat("a", 100), strings.Repeat("a", 100), false},
		{"101 chars inválido", strings.Repeat("a", 101), "", true},
		{"vazio", "", "", true},
		{"apenas espaços", "   ", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewCategoryName(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryName)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}

func TestCategoryNameEquality(t *testing.T) {
	t.Parallel()
	a, _ := vos.NewCategoryName("X")
	b, _ := vos.NewCategoryName(" X ")
	assert.Equal(t, a, b)
}
