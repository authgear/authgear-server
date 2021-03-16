package exactmatchlist_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/util/exactmatchlist"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExactMatchList(t *testing.T) {
	Convey("ExactMatchList", t, func() {
		data := `
			example.com
			testing.com

			FREE-MAIL.COM
		`
		Convey("should skip empty line in list", func() {
			list, err := exactmatchlist.New(data, false)
			So(err, ShouldBeNil)
			So(list.NumEntries(), ShouldEqual, 3)
		})

		Convey("should exact match entries", func() {
			list, _ := exactmatchlist.New(data, false)
			cases := []struct {
				input  string
				result bool
			}{
				{"example.com", true},
				{"testing.com", true},
				{"FREE-MAIL.COM", true},
				{"free-mail.COM", false},
				{"freemail.com", false},
				{"anything", false},
			}

			for _, c := range cases {
				result, _ := list.Matched(c.input)
				So(result, ShouldEqual, c.result)
			}
		})

		Convey("should exact match fold case entries", func() {
			list, _ := exactmatchlist.New(data, true)
			cases := []struct {
				input  string
				result bool
			}{
				{"example.com", true},
				{"testing.com", true},
				{"FREE-MAIL.COM", true},
				{"free-mail.COM", true},
				{"free-mail.com", true},
				{"freemail.com", false},
				{"anything", false},
			}

			for _, c := range cases {
				result, _ := list.Matched(c.input)
				So(result, ShouldEqual, c.result)
			}
		})
	})
}
