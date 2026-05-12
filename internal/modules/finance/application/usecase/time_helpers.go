package usecase

import "time"

// addMonthsClamped returns base shifted forward by the given number of months,
// clamping the day to the last valid day of the target month.
//
// Plain time.Time.AddDate(0, n, 0) normalises overflowing days (e.g. Jan 31 + 1mo
// → Mar 3) which can cause installments of card purchases made at the end of the
// month to skip a billing cycle entirely (RF-04 + RF-15). Clamping the day to the
// last day of the target month preserves one installment per consecutive cycle.
func addMonthsClamped(base time.Time, months int) time.Time {
	year, month := normaliseYearMonth(base.Year(), int(base.Month())+months)
	loc := base.Location()
	day := base.Day()
	if last := lastDayOfMonth(year, time.Month(month), loc); day > last {
		day = last
	}
	return time.Date(year, time.Month(month), day,
		base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), loc)
}

func normaliseYearMonth(year, month int) (int, int) {
	for month > 12 {
		month -= 12
		year++
	}
	for month < 1 {
		month += 12
		year--
	}
	return year, month
}

func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}
