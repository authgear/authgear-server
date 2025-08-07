package timeutil

import "time"

func Last30Days(now time.Time) (time.Time, time.Time) {
	rangeTo := now
	rangeFrom := now.AddDate(0, 0, -30)
	return rangeFrom, rangeTo
}
