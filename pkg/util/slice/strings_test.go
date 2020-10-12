package slice

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExceptStrings(t *testing.T) {
	Convey("ExceptStrings", t, func() {
		Convey("return new slice without unwanted items", func() {
			result := ExceptStrings([]string{
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
			result := ExceptStrings([]string{
				"1",
				"2",
				"3",
			}, []string{
				"4",
			})
			So(len(result), ShouldEqual, 3)
		})

		Convey("works with duplicated items to filter", func() {
			result := ExceptStrings([]string{
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

func TestContainsString(t *testing.T) {
	Convey("ContainsString", t, func() {
		Convey("should check if a string not in a slice", func() {
			result := ContainsString(
				[]string{
					"1",
					"2",
					"3",
				},
				"4",
			)
			So(result, ShouldEqual, false)
		})

		Convey("should check if a string in a slice", func() {
			result := ContainsString(
				[]string{
					"1",
					"2",
					"3",
				},
				"1",
			)
			So(result, ShouldEqual, true)
		})

	})
}

func TestAppendIfUniqueStrings(t *testing.T) {
	Convey("AppendIfUniqueStrings", t, func() {
		So(AppendIfUniqueStrings(nil, ""), ShouldResemble, []string{""})
		So(AppendIfUniqueStrings([]string{""}, ""), ShouldResemble, []string{""})
		So(AppendIfUniqueStrings([]string{""}, "a"), ShouldResemble, []string{"", "a"})
		So(AppendIfUniqueStrings([]string{"a"}, "a"), ShouldResemble, []string{"a"})
		So(AppendIfUniqueStrings([]string{"a"}, "b"), ShouldResemble, []string{"a", "b"})
	})
}
