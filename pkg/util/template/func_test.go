package template

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMakeMap(t *testing.T) {
	Convey("MakeMap", t, func() {
		So(func() { MakeMap("key") }, ShouldPanic)
		So(func() { MakeMap(1, 2) }, ShouldPanic)
		So(MakeMap("key", 1), ShouldResemble, map[string]interface{}{
			"key": 1,
		})
	})
}
