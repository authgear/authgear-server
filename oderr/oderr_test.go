package oderr

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewError(t *testing.T) {
	Convey("An Error", t, func() {
		err := New(1, "some message")

		Convey("returns code correctly", func() {
			So(err.Code(), ShouldEqual, 1)
		})

		Convey("returns message correctly", func() {
			So(err.Message(), ShouldEqual, "some message")
		})

		Convey("Error()s in format {code}: {message}", func() {
			So(err.Error(), ShouldEqual, "1: some message")
		})
	})
}
