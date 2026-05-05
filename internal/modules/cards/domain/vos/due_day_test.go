package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDueDay(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"1 válido", 1, false},
		{"31 válido", 31, false},
		{"10 válido", 10, false},
		{"0 inválido", 0, true},
		{"32 inválido", 32, true},
		{"-1 inválido", -1, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.NewDueDay(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, vos.ErrInvalidDueDay)
				assert.Equal(t, 0, got.Int())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.input, got.Int())
		})
	}
}
