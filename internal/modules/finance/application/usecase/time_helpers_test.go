package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BUG-004 regression: addMonthsClamped must NOT normalise day overflow the way
// time.AddDate(0, n, 0) does. For a purchase made on 31/Jan with closing day 1,
// time.AddDate would push installment #2 from Feb to "Mar 3", causing the
// invoice assigner to skip Feb entirely. Clamping the day preserves one
// installment per consecutive cycle (RF-04/RF-15).
func TestAddMonthsClamped_PreservesEndOfMonth(t *testing.T) {
	t.Parallel()
	loc := time.UTC
	base := time.Date(2026, time.January, 31, 10, 0, 0, 0, loc)

	tests := []struct {
		name       string
		months     int
		wantYear   int
		wantMonth  time.Month
		wantDay    int
		altWantDay int // optional alternative day for months with 30 days
	}{
		{name: "+0 months keeps the same date", months: 0, wantYear: 2026, wantMonth: time.January, wantDay: 31},
		{name: "+1 month clamps to last day of Feb (2026 non-leap → 28)", months: 1, wantYear: 2026, wantMonth: time.February, wantDay: 28},
		{name: "+2 months keeps 31 in March", months: 2, wantYear: 2026, wantMonth: time.March, wantDay: 31},
		{name: "+3 months clamps to 30 in April", months: 3, wantYear: 2026, wantMonth: time.April, wantDay: 30},
		{name: "+4 months keeps 31 in May", months: 4, wantYear: 2026, wantMonth: time.May, wantDay: 31},
		{name: "+12 months wraps year", months: 12, wantYear: 2027, wantMonth: time.January, wantDay: 31},
		{name: "+13 months wraps and clamps", months: 13, wantYear: 2027, wantMonth: time.February, wantDay: 28},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := addMonthsClamped(base, tc.months)
			assert.Equal(t, tc.wantYear, got.Year())
			assert.Equal(t, tc.wantMonth, got.Month())
			assert.Equal(t, tc.wantDay, got.Day())
			// time-of-day and location must survive the shift
			assert.Equal(t, base.Hour(), got.Hour())
			assert.Equal(t, base.Minute(), got.Minute())
			assert.Equal(t, base.Location(), got.Location())
		})
	}
}

// BUG-004 cross-check: AddDate normalises Jan 31 + 1mo → Mar 3 whereas the new
// helper produces Feb 28 — pinning the regression in a single assertion.
func TestAddMonthsClamped_DivergesFromAddDateForEndOfMonth(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)

	normalised := base.AddDate(0, 1, 0)
	clamped := addMonthsClamped(base, 1)

	// Sanity: AddDate rolls over into March.
	assert.Equal(t, time.March, normalised.Month())
	// Helper stays in February.
	assert.Equal(t, time.February, clamped.Month())
	assert.Equal(t, 28, clamped.Day())
}

// BUG-004: leap year — Feb has 29 days in 2024.
func TestAddMonthsClamped_LeapYearFebruary(t *testing.T) {
	t.Parallel()
	base := time.Date(2024, time.January, 30, 12, 30, 45, 0, time.UTC)
	got := addMonthsClamped(base, 1)
	assert.Equal(t, time.February, got.Month())
	assert.Equal(t, 29, got.Day())
}
