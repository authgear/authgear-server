package timeutil

import (
	"time"
)

func FirstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 1, 1, 0, 0, 0, 0, timezone)
	// ensure the checking starts from the date which is before the result date
	// so use `week - 1`
	date = date.AddDate(0, 0, (week-1)*7)
	isoYear, isoWeek := date.ISOWeek()

	// get the nearest monday
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}

	// for the case that the date becomes last year
	// move forward to the first day of the first week
	for isoYear < year {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	// move forward to the inputted week
	for isoWeek < week {
		date = date.AddDate(0, 0, 7)
		_, isoWeek = date.ISOWeek()
	}
	return date
}
