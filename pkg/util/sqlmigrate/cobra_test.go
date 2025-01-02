package sqlmigrate

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCobraParseMigrateUpArgs(t *testing.T) {
	Convey("CobraParseMigrateUpArgs", t, func() {
		var n int
		var err error

		var f = CobraParseMigrateUpArgs

		// No args mean migrate to latest.
		n, err = f(nil)
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 0)

		// Accept integer only.
		_, err = f([]string{"nonsense"})
		So(err, ShouldNotBeNil)

		// Accept positive integer only.
		_, err = f([]string{"0"})
		So(err, ShouldBeError, "n must be a positive integer: 0")

		n, err = f([]string{"1"})
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 1)
	})
}

func TestCobraParseMigrateDownArgs(t *testing.T) {
	Convey("CobraParseMigrateDownArgs", t, func() {
		var n int
		var err error

		var f = CobraParseMigrateDownArgs

		// Requires args
		_, err = f(nil)
		So(err, ShouldBeError, "n must be a positive integer")

		// Accept integer only.
		_, err = f([]string{"nonsense"})
		So(err, ShouldNotBeNil)

		// Accept positive integer only.
		_, err = f([]string{"0"})
		So(err, ShouldBeError, "n must be a positive integer: 0")

		n, err = f([]string{"1"})
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 1)

		// Accept all as a special case
		n, err = f([]string{"all"})
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 0)
	})
}
