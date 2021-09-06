package timeutil

import (
	"fmt"
	"time"
)

func FirstDayOfISOWeek(year int, week int, timezone *time.Location) (*time.Time, error) {
	// The first week must contain 4 January.
	// https://en.wikipedia.org/wiki/ISO_week_date#First_week
	_, firstWeek := time.Date(year, 1, 4, 0, 0, 0, 0, timezone).ISOWeek()
	// The last week must contain 28 December.
	// https://en.wikipedia.org/wiki/ISO_week_date#Last_week
	_, lastWeek := time.Date(year, 12, 28, 0, 0, 0, 0, timezone).ISOWeek()
	if week < firstWeek || week > lastWeek {
		return nil, fmt.Errorf("invalid week: %vW%v not in [%v, %v]", year, week, firstWeek, lastWeek)
	}

	date := time.Date(year, 1, 4, 0, 0, 0, 0, timezone)
	// get the first Monday of the year
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	date = date.AddDate(0, 0, (week-1)*7)
	return &date, nil
}
