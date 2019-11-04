package validation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJSONPointer(t *testing.T) {
	Convey("JSONPointer", t, func() {
		So(JSONPointer("entries", 0, "name"), ShouldEqual, "/entries/0/name")
		So(JSONPointer("http://example.com", 1, "~"), ShouldEqual, "/http:~1~1example.com/1/~0")
	})
}
