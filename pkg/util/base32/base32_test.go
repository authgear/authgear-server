package base32

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNormalization(t *testing.T) {
	Convey("Normalize", t, func() {
		var result string
		var err error

		Convey("should normalize correctly", func() {
			result, err = Normalize("")
			So(result, ShouldEqual, "")
			So(err, ShouldBeNil)

			result, err = Normalize("Op7d0V61")
			So(result, ShouldEqual, "0P7D0V61")
			So(err, ShouldBeNil)

			result, err = Normalize("1iO0oLIl")
			So(result, ShouldEqual, "11000111")
			So(err, ShouldBeNil)

			result, err = Normalize("IV7R6m9vQoiHTwFeGkQg5nLVFasYA2OUoBe6soF4z")
			So(result, ShouldEqual, "1V7R6M9VQ01HTWFEGKQG5N1VFASYA20U0BE6S0F4Z")
			So(err, ShouldBeNil)
		})

		Convey("should remove separators", func() {
			result, err = Normalize("---")
			So(result, ShouldEqual, "")
			So(err, ShouldBeNil)

			result, err = Normalize("-a-O-4-Q-")
			So(result, ShouldEqual, "A04Q")
			So(err, ShouldBeNil)

			result, err = Normalize("gA49-0ikL-maQO-EWMe")
			So(result, ShouldEqual, "GA4901K1MAQ0EWME")
			So(err, ShouldBeNil)
		})

		Convey("should fail for invalid characters", func() {
			result, err = Normalize(".")
			So(result, ShouldEqual, "")
			So(err, ShouldBeError, InvalidBase32Character('.'))

			result, err = Normalize("iXqgbq+IV7R6m9")
			So(result, ShouldEqual, "")
			So(err, ShouldBeError, InvalidBase32Character('+'))
		})
	})
}
