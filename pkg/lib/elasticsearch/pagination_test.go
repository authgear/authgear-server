package elasticsearch

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestCursorToSearchAfter(t *testing.T) {
	Convey("CursorToSearchAfter", t, func() {
		actual, err := CursorToSearchAfter("")
		So(err, ShouldBeNil)
		So(actual, ShouldBeNil)

		actual, err = CursorToSearchAfter(model.PageCursor("WzEyM10"))
		So(err, ShouldBeNil)
		So(actual, ShouldResemble, []interface{}{123.0})
	})
}

func TestSortToCursor(t *testing.T) {
	Convey("SortToCursor", t, func() {
		actual, err := SortToCursor(nil)
		So(err, ShouldBeNil)
		So(actual, ShouldEqual, model.PageCursor(""))

		actual, err = SortToCursor([]interface{}{123})
		So(err, ShouldBeNil)
		So(actual, ShouldEqual, model.PageCursor("WzEyM10"))
	})
}
