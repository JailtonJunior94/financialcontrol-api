package vos

import (
	"time"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
)

// Period represents a calendar month in a specific timezone (America/Sao_Paulo by default).
// Covers RF-20/RF-51.
type Period struct {
	year     int
	month    time.Month
	location *time.Location
}

// NewPeriod constructs a Period. year must be >= 1 and month in [1, 12].
// loc is optional; when nil, America/Sao_Paulo is used.
func NewPeriod(year int, month int, loc *time.Location) (Period, error) {
	if year < 1 || month < 1 || month > 12 {
		return Period{}, domain.ErrInvalidYearMonth
	}
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("America/Sao_Paulo")
		if err != nil {
			return Period{}, err
		}
	}
	return Period{year: year, month: time.Month(month), location: loc}, nil
}

// Bounds returns the half-open interval [start, end) in UTC equivalent to the
// month in the Period's timezone.
//
// Example: Period{2026, 5, "America/Sao_Paulo"}.Bounds() returns
//
//	[2026-05-01T03:00:00Z, 2026-06-01T03:00:00Z)
//
// because America/Sao_Paulo is UTC-3 in May (no DST).
func (p Period) Bounds() (time.Time, time.Time) {
	start := time.Date(p.year, p.month, 1, 0, 0, 0, 0, p.location)
	end := start.AddDate(0, 1, 0)
	return start.UTC(), end.UTC()
}

func (p Period) Year() int                { return p.year }
func (p Period) Month() time.Month        { return p.month }
func (p Period) Location() *time.Location { return p.location }
