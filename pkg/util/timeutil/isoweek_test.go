package timeutil_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func TestFirstDayOfISOWeek(t *testing.T) {
	Convey("FirstDayOfISOWeek", t, func() {

		Convey("should convert valid week number", func() {
			// Test all weeks from year 1900 to 3000
			t := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
			// start from the first monday
			for t.Weekday() != time.Monday {
				t = t.AddDate(0, 0, 1)
			}
			for t.Year() < 3000 {
				iosYear, isoWeek := t.ISOWeek()
				calculatedDate, _ := timeutil.FirstDayOfISOWeek(iosYear, isoWeek, time.UTC)
				So(calculatedDate, ShouldResemble, &t)
				t = t.AddDate(0, 0, 7)
			}
		})

		Convey("should return error for invalid week number", func() {
			_, err := timeutil.FirstDayOfISOWeek(2021, 53, time.UTC)
			So(err, ShouldBeError, "invalid week: 2021W53 not in [1, 52]")
		})
	})
}
