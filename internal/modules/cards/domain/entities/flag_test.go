package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
)

func TestNewFlag(t *testing.T) {
	t.Parallel()

	flagID, err := vos.ParseFlagID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	assert.NoError(t, err)

	cases := []struct {
		name   string
		id     vos.FlagID
		fName  string
		active bool
	}{
		{"active flag", flagID, "Visa", true},
		{"inactive flag", flagID, "Mastercard", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := entities.NewFlag(tc.id, tc.fName, tc.active)
			assert.Equal(t, tc.id, f.ID())
			assert.Equal(t, tc.fName, f.Name())
			assert.Equal(t, tc.active, f.Active())
		})
	}
}
