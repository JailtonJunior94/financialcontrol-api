package vos_test

import (
	"strings"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCardNumber(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"12 dígitos válido", strings.Repeat("1", 12), false},
		{"19 dígitos válido", strings.Repeat("1", 19), false},
		{"16 dígitos válido", "4111111111111111", false},
		{"11 dígitos inválido", strings.Repeat("1", 11), true},
		{"20 dígitos inválido", strings.Repeat("1", 20), true},
		{"com letras inválido", "4111111111111a11", true},
		{"com espaços inválido", "4111 1111 1111 1111", true},
		{"string vazia", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewCardNumber(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, vos.ErrInvalidCardNumber)
				assert.Empty(t, got.String())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.input, got.String())
		})
	}
}
