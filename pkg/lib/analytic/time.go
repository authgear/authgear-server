package analytic

import (
	"time"

	periodicalutil "github.com/authgear/authgear-server/pkg/util/periodical"
)

func GetDateListByRangeInclusive(rangeFrom time.Time, rangeTo time.Time, periodical periodicalutil.Type) []time.Time {
	dateList := []time.Time{}
	date := rangeFrom.UTC()
	switch periodical {
	case periodicalutil.Monthly:
		date = time.Date(rangeFrom.Year(), rangeFrom.Month(), 1, 0, 0, 0, 0, time.UTC)
		if rangeFrom.Day() != 1 {
			date = date.AddDate(0, 1, 0)
		}
		for {
			// Termination
			if date.After(rangeTo) {
				break
			}
			dateList = append(dateList, date)
			date = date.AddDate(0, 1, 0)
		}
	case periodicalutil.Weekly:
		for date.Weekday() != time.Monday {
			date = date.AddDate(0, 0, 1)
		}
		for {
			// Termination
			if date.After(rangeTo) {
				break
			}
			dateList = append(dateList, date)
			date = date.AddDate(0, 0, 7)
		}
	case periodicalutil.Daily:
		date = time.Date(rangeFrom.Year(), rangeFrom.Month(), rangeFrom.Day(), 0, 0, 0, 0, time.UTC)
		for {
			// Termination
			if date.After(rangeTo) {
				break
			}
			dateList = append(dateList, date)
			date = date.AddDate(0, 0, 1)
		}
	}
	return dateList
}
