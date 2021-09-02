package timeutil_test

import (
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/util/timeutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFirstDayOfISOWeek(t *testing.T) {
	Convey("FirstDayOfISOWeek", t, func() {
		// Test all weeks from year 1900 to 3000
		t := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
		// start from the first monday
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, 1)
		}
		for t.Year() < 3000 {
			iosYear, isoWeek := t.ISOWeek()
			calculatedDate := timeutil.FirstDayOfISOWeek(iosYear, isoWeek, time.UTC)
			So(calculatedDate, ShouldEqual, t)
			t = t.AddDate(0, 0, 7)
		}
	})
}
