package blocklist_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/blocklist"
)

func TestBlocklist(t *testing.T) {
	Convey("Blocklist", t, func() {
		data := `
			# comment here
			adm1n

			/www\d*/
			!www01
		`
		Convey("should parse input data", func() {
			list, err := blocklist.New(data)
			So(err, ShouldBeNil)
			So(list.NumEntries(), ShouldEqual, 3)

			_, err = blocklist.New(`/\c/`)
			So(err, ShouldBeError, "invalid blocklist entry at line 1: error parsing regexp: invalid escape sequence: `\\c`")
		})

		Convey("should match entries", func() {
			list, _ := blocklist.New(data)

			So(list.IsBlocked("test"), ShouldBeFalse)
			So(list.IsBlocked("admin"), ShouldBeFalse)
			So(list.IsBlocked("adm1n"), ShouldBeTrue)
			So(list.IsBlocked("www"), ShouldBeTrue)
			So(list.IsBlocked("www01"), ShouldBeFalse)
			So(list.IsBlocked("www02"), ShouldBeTrue)
			So(list.IsBlocked("www03"), ShouldBeTrue)
		})
	})
}
