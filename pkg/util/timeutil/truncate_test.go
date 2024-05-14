package timeutil_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func TestDateUtil(t *testing.T) {
	Convey("TestDateUtil", t, func() {

		Convey("TruncateToDate", func() {
			So(
				timeutil.TruncateToDate(time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)),
				ShouldResemble,
				time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			)
		})

		Convey("FirstDayOfTheMonth", func() {
			So(
				timeutil.FirstDayOfTheMonth(time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)),
				ShouldResemble,
				time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
			)
		})

		Convey("MondayOfTheWeek", func() {
			So(
				timeutil.MondayOfTheWeek(time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)),
				ShouldResemble,
				time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			)
			So(
				timeutil.MondayOfTheWeek(time.Date(2006, 1, 7, 3, 4, 5, 6, time.UTC)),
				ShouldResemble,
				time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			)
		})
	})
}
