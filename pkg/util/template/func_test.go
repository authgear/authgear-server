package template

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRFC3339(t *testing.T) {
	date := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)

	Convey("RFC3339", t, func() {
		Convey("it supports time.Time", func() {
			So(RFC3339(date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it supports *time.Time", func() {
			So(RFC3339(&date), ShouldEqual, "2006-01-02T03:04:05Z")
		})
		Convey("it does not fail for other data type", func() {
			So(RFC3339(nil), ShouldEqual, "INVALID_DATE")
			So(RFC3339(false), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(0.0), ShouldEqual, "INVALID_DATE")
			So(RFC3339(""), ShouldEqual, "INVALID_DATE")
			So(RFC3339(struct{}{}), ShouldEqual, "INVALID_DATE")
			So(RFC3339([]struct{}{}), ShouldEqual, "INVALID_DATE")
		})
	})

	Convey("IsNil", t, func() {
		So(IsNil(nil), ShouldBeTrue)

		var p *int64
		So(IsNil(p), ShouldBeTrue)

		var v int64
		p = &v
		So(IsNil(p), ShouldBeFalse)
		So(IsNil(v), ShouldBeFalse)

		p = nil
		So(IsNil(p), ShouldBeTrue)
	})

	Convey("ShowAttributeValue", t, func() {
		newString := func(s string) *string {
			return &s
		}

		newInt := func(i int64) *int64 {
			return &i
		}

		newFloat := func(f float64) *float64 {
			return &f
		}

		So(ShowAttributeValue(nil), ShouldEqual, "")
		So(ShowAttributeValue(1), ShouldEqual, "1")
		So(ShowAttributeValue(1.2), ShouldEqual, "1.2")
		So(ShowAttributeValue(100000000), ShouldEqual, "100000000")
		So(ShowAttributeValue(100000000.002), ShouldEqual, "100000000.002")
		So(ShowAttributeValue("test"), ShouldEqual, "test")

		var ip *int64
		So(ShowAttributeValue(ip), ShouldEqual, "")

		ip = newInt(100000000)
		So(ShowAttributeValue(ip), ShouldEqual, "100000000")

		var f32 float32 = 0.00002
		So(ShowAttributeValue(f32), ShouldEqual, "0.00002")

		var f64 float64 = 0.00002
		So(ShowAttributeValue(f64), ShouldEqual, "0.00002")

		var fp *float64
		So(ShowAttributeValue(fp), ShouldEqual, "")

		fp = newFloat(0)
		So(ShowAttributeValue(fp), ShouldEqual, "0")

		fp = newFloat(100000000.01)
		So(ShowAttributeValue(fp), ShouldEqual, "100000000.01")

		var sp *string
		So(ShowAttributeValue(sp), ShouldEqual, "")

		sp = newString("test")
		So(ShowAttributeValue(sp), ShouldEqual, "test")

	})
}
