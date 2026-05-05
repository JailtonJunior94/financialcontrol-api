package vos_test

import (
	"strings"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCardName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"nome válido", "Nubank", "Nubank", false},
		{"nome com espaços nas bordas", "  Itaú  ", "Itaú", false},
		{"100 chars válido", strings.Repeat("a", 100), strings.Repeat("a", 100), false},
		{"101 chars inválido", strings.Repeat("a", 101), "", true},
		{"string vazia", "", "", true},
		{"apenas espaços", "   ", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewCardName(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, vos.ErrInvalidCardName)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}
