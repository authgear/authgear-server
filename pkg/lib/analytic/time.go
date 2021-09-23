package analytic

import (
	"time"

	periodicalutil "github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func GetDateListByRangeInclusive(rangeFrom time.Time, rangeTo time.Time, periodical periodicalutil.Type) []time.Time {
	dateList := []time.Time{}
	date := rangeFrom.UTC()
	switch periodical {
	case periodicalutil.Monthly:
		date = timeutil.FirstDayOfTheMonth(rangeFrom)
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
		date = timeutil.TruncateToDate(rangeFrom)
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
