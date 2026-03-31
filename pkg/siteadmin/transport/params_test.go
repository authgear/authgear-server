package transport

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetOptionalIntParam(t *testing.T) {
	Convey("getOptionalIntParam", t, func() {
		Convey("returns nil when param is absent", func() {
			q := url.Values{}
			v, err := getOptionalIntParam(q, "count")
			So(err, ShouldBeNil)
			So(v, ShouldBeNil)
		})

		Convey("returns parsed value when param is present", func() {
			q := url.Values{"count": {"42"}}
			v, err := getOptionalIntParam(q, "count")
			So(err, ShouldBeNil)
			So(v, ShouldNotBeNil)
			So(*v, ShouldEqual, 42)
		})

		Convey("returns error when param is not an integer", func() {
			q := url.Values{"count": {"abc"}}
			v, err := getOptionalIntParam(q, "count")
			So(err, ShouldNotBeNil)
			So(v, ShouldBeNil)
		})
	})
}

func TestGetIntParam(t *testing.T) {
	Convey("getIntParam", t, func() {
		Convey("returns error when param is absent", func() {
			q := url.Values{}
			_, err := getIntParam(q, "count")
			So(err, ShouldNotBeNil)
		})

		Convey("returns parsed value when param is present", func() {
			q := url.Values{"count": {"7"}}
			v, err := getIntParam(q, "count")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, 7)
		})

		Convey("returns error when param is not an integer", func() {
			q := url.Values{"count": {"xyz"}}
			_, err := getIntParam(q, "count")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetOptionalDateParam(t *testing.T) {
	Convey("getOptionalDateParam", t, func() {
		Convey("returns nil when param is absent", func() {
			q := url.Values{}
			v, err := getOptionalDateParam(q, "date")
			So(err, ShouldBeNil)
			So(v, ShouldBeNil)
		})

		Convey("returns value when param is a valid date", func() {
			q := url.Values{"date": {"2024-03-15"}}
			v, err := getOptionalDateParam(q, "date")
			So(err, ShouldBeNil)
			So(v, ShouldNotBeNil)
			So(*v, ShouldEqual, "2024-03-15")
		})

		Convey("returns error when param is not a valid date", func() {
			q := url.Values{"date": {"15-03-2024"}}
			v, err := getOptionalDateParam(q, "date")
			So(err, ShouldNotBeNil)
			So(v, ShouldBeNil)
		})
	})
}

func TestGetDateParam(t *testing.T) {
	Convey("getDateParam", t, func() {
		Convey("returns error when param is absent", func() {
			q := url.Values{}
			_, err := getDateParam(q, "date")
			So(err, ShouldNotBeNil)
		})

		Convey("returns value when param is a valid date", func() {
			q := url.Values{"date": {"2024-01-01"}}
			v, err := getDateParam(q, "date")
			So(err, ShouldBeNil)
			So(v, ShouldEqual, "2024-01-01")
		})

		Convey("returns error when param is not a valid date", func() {
			q := url.Values{"date": {"not-a-date"}}
			_, err := getDateParam(q, "date")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestValidateMonth(t *testing.T) {
	Convey("validateMonth", t, func() {
		Convey("accepts values 1 through 12", func() {
			for m := 1; m <= 12; m++ {
				So(validateMonth("month", m), ShouldBeNil)
			}
		})

		Convey("rejects 0", func() {
			So(validateMonth("month", 0), ShouldNotBeNil)
		})

		Convey("rejects 13", func() {
			So(validateMonth("month", 13), ShouldNotBeNil)
		})

		Convey("rejects negative values", func() {
			So(validateMonth("month", -1), ShouldNotBeNil)
		})
	})
}
