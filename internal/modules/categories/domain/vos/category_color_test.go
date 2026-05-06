package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategoryColor(t *testing.T) {
	t.Parallel()
	valid := []string{"red", "orange", "yellow", "green", "teal", "blue", "indigo", "purple", "pink", "brown", "gray"}
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"upper case normalizado", "RED", "red", false},
		{"trim", "  blue  ", "blue", false},
		{"vazio", "", "", true},
		{"valor não permitido", "magenta", "", true},
	}
	for _, c := range valid {
		cases = append(cases, struct {
			name    string
			input   string
			want    string
			wantErr bool
		}{name: "válido " + c, input: c, want: c, wantErr: false})
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewCategoryColor(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrInvalidCategoryColor)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}
