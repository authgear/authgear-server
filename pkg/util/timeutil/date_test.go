package timeutil_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func TestDate(t *testing.T) {
	Convey("TestDate", t, func() {
		Convey("test Decode", func() {
			d := timeutil.Date{}
			e := d.Decode("2006-01-02")
			So(e, ShouldBeNil)
			So(d, ShouldResemble, timeutil.Date(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)))
		})

		Convey("test MarshalJSON", func() {
			t := timeutil.Date(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
			dateB, _ := json.Marshal(&t)
			So(string(dateB), ShouldResemble, `"2006-01-02"`)

			var tPtr *time.Time
			dateB, _ = json.Marshal(tPtr)
			So(string(dateB), ShouldResemble, `null`)
		})
	})
}
