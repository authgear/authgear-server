package matchlist_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/matchlist"
)

func TestMatchList(t *testing.T) {
	Convey("MatchList", t, func() {
		data := `
			testing-sample1
			testing-sample2
			TESTING-sample3
			TESTING-SAMPLE4
		`
		Convey("should skip empty line in list", func() {
			list, err := matchlist.New(data, false, false)
			So(err, ShouldBeNil)
			So(list.NumEntries(), ShouldEqual, 4)
		})

		Convey("should exact match entries", func() {
			list, _ := matchlist.New(data, false, false)
			cases := []struct {
				input  string
				result bool
			}{
				{"testing-sample1", true},
				{"testing-sample2", true},
				{"TESTING-sample3", true},
				{"TESTING-SAMPLE4", true},
				{"testing-SAMPLE1", false},
				{"TESTING-sample2", false},
				{"testing-sample3", false},
				{"testing-sample4", false},
				{"anything", false},
			}

			for _, c := range cases {
				result, _ := list.Matched(c.input)
				So(result, ShouldEqual, c.result)
			}
		})

		Convey("should exact match fold case entries", func() {
			list, _ := matchlist.New(data, true, false)
			cases := []struct {
				input  string
				result bool
			}{
				{"testing-sample1", true},
				{"testing-sample2", true},
				{"TESTING-sample3", true},
				{"TESTING-SAMPLE4", true},
				{"testing-SAMPLE1", true},
				{"TESTING-sample2", true},
				{"testing-sample3", true},
				{"testing-sample4", true},
				{"Extra-testing-SAMPLE1-test", false},
				{"testing-sample4-test", false},
				{"anything", false},
			}

			for _, c := range cases {
				result, _ := list.Matched(c.input)
				So(result, ShouldEqual, c.result)
			}
		})

		Convey("should match string contain", func() {
			list, _ := matchlist.New(data, true, true)
			cases := []struct {
				input  string
				result bool
			}{
				{"testing-sample1", true},
				{"testing-sample2", true},
				{"TESTING-sample3", true},
				{"TESTING-SAMPLE4", true},
				{"testing-SAMPLE1", true},
				{"TESTING-sample2", true},
				{"testing-sample3", true},
				{"testing-sample4", true},
				{"Extra-testing-SAMPLE1-test", true},
				{"testing-sample4-test", true},
				{"anything", false},
			}

			for _, c := range cases {
				result, _ := list.Matched(c.input)
				So(result, ShouldEqual, c.result)
			}
		})
	})
}
