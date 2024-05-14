package analytic_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

func TestGetDateListByRangeInclusive(t *testing.T) {
	Convey("GetDateListByRangeInclusive", t, func() {
		list := analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
			periodical.Monthly,
		)
		So(len(list), ShouldEqual, 1)

		list = analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			periodical.Monthly,
		)
		So(len(list), ShouldEqual, 0)

		list = analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC),
			periodical.Monthly,
		)
		So(len(list), ShouldEqual, 12)
		So(list[0], ShouldResemble, time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC))
		So(list[len(list)-1], ShouldResemble, time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC))

		list = analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC),
			periodical.Weekly,
		)
		So(len(list), ShouldEqual, 53)
		So(list[0], ShouldResemble, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
		So(list[len(list)-1], ShouldResemble, time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC))

		list = analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 6, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 1, 6, 0, 0, 0, 0, time.UTC),
			periodical.Weekly,
		)
		So(len(list), ShouldEqual, 52)
		So(list[0], ShouldResemble, time.Date(2006, 1, 9, 0, 0, 0, 0, time.UTC))
		So(list[len(list)-1], ShouldResemble, time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC))

		list = analytic.GetDateListByRangeInclusive(
			time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC),
			periodical.Daily,
		)
		So(len(list), ShouldEqual, 366)
		So(list[0], ShouldResemble, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
		So(list[len(list)-1], ShouldResemble, time.Date(2007, 1, 2, 0, 0, 0, 0, time.UTC))
	})
}
