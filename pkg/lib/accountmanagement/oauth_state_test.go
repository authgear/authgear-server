package accountmanagement

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractStateFromQuery(t *testing.T) {
	Convey("ExtractStateFromQuery", t, func() {
		state, err := ExtractStateFromQuery("")
		So(err, ShouldBeNil)
		So(state, ShouldEqual, "")

		state, err = ExtractStateFromQuery("?code=code")
		So(err, ShouldBeNil)
		So(state, ShouldEqual, "")

		state, err = ExtractStateFromQuery("code=code")
		So(err, ShouldBeNil)
		So(state, ShouldEqual, "")

		state, err = ExtractStateFromQuery("?code=code&state=state")
		So(err, ShouldBeNil)
		So(state, ShouldEqual, "state")

		state, err = ExtractStateFromQuery("code=code&state=state")
		So(err, ShouldBeNil)
		So(state, ShouldEqual, "state")
	})
}
