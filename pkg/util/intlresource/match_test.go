package intlresource

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/intl"
)

type testItem string

func (i testItem) GetLanguageTag() string {
	return string(i)
}

func TestMatch(t *testing.T) {
	Convey("Match", t, func() {
		var matched LanguageItem
		var err error

		// Match preferred if possible
		matched, err = Match([]string{"zh"}, "ja", []LanguageItem{
			testItem("zh"),
			testItem("ja"),
			testItem(intl.DefaultLanguage),
		})
		So(err, ShouldBeNil)
		So(matched, ShouldEqual, testItem("zh"))

		// Match fallback if possible
		matched, err = Match([]string{"zh"}, "ja", []LanguageItem{
			testItem("ja"),
			testItem(intl.DefaultLanguage),
		})
		So(err, ShouldBeNil)
		So(matched, ShouldEqual, testItem("ja"))
	})
}
