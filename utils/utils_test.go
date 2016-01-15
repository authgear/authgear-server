package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStrSliceWithout(t *testing.T) {
	Convey("StrSliceWithout", t, func() {
		Convey("return new slice without unwanted items", func() {
			result := StrSliceWithout([]string{
				"1",
				"2",
				"3",
			}, []string{
				"1",
				"3",
			})
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "2")
		})

		Convey("should return all items if no items is filtered", func() {
			result := StrSliceWithout([]string{
				"1",
				"2",
				"3",
			}, []string{
				"4",
			})
			So(len(result), ShouldEqual, 3)
		})

		Convey("works with duplicated items to filter", func() {
			result := StrSliceWithout([]string{
				"1",
				"2",
				"3",
				"4",
				"5",
				"6",
				"7",
				"8",
				"9",
			}, []string{
				"4",
				"4",
				"1",
				"2",
				"3",
				"1",
				"2",
				"3",
				"7",
				"8",
				"9",
			})
			So(len(result), ShouldEqual, 2)
		})
	})

}
