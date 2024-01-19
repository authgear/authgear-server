package intl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHTMLDir(t *testing.T) {
	Convey("HTMLDir", t, func() {
		Convey("report ltr most of the case", func() {
			ltr := []string{
				"en",
				"zh",
				"zh-Hant",
				"zh-Hans",
				"ja",
				"az",
			}

			for _, tag := range ltr {
				So(HTMLDir(tag), ShouldEqual, "ltr")
			}
		})

		Convey("report rtl", func() {
			rtl := []string{
				"apc",
				"ar",
				"az-Arab",
				"bm-Nkoo",
				"ff-Adlm",
			}

			for _, tag := range rtl {
				So(HTMLDir(tag), ShouldEqual, "rtl")
			}
		})
	})
}
