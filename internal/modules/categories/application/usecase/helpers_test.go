package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func newFixedClock() fixedClock {
	return fixedClock{now: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)}
}

func mustUserID(t *testing.T) identityvo.UserID {
	t.Helper()
	id, err := identityvo.ParseUserID("11111111-1111-4111-8111-111111111111")
	require.NoError(t, err)
	return id
}
