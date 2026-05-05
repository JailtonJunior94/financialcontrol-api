package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlagID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"UUID válido", "550e8400-e29b-41d4-a716-446655440000", false},
		{"string vazia", "", true},
		{"UUID inválido", "not-a-uuid", true},
		{"UUID incompleto", "550e8400-e29b-41d4", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := vos.ParseFlagID(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, vos.ErrInvalidFlagID)
				assert.Empty(t, got.String())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.input, got.String())
		})
	}
}
