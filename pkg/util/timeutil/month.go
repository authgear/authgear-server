package timeutil

import (
	"time"
)

// PreviousMonth returns YYYY-MM of t, where YYYY-MM is the previous month.
func PreviousMonth(t time.Time) (year int, month time.Month) {
	// Set to the first day of the month
	// This avoids you wonder `2025-03-31 - 1month = 2025-02-31`
	t = FirstDayOfTheMonth(t)
	// Minus 1 month
	t = t.AddDate(0, -1, 0)
	year = t.Year()
	month = t.Month()
	return
}
