package nameutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormat(t *testing.T) {
	Convey("Format", t, func() {
		So(Format("", "", ""), ShouldEqual, "")
		So(Format("John", "", ""), ShouldEqual, "John")
		So(Format("John", "", "Doe"), ShouldEqual, "John Doe")
		So(Format("", "", "Doe"), ShouldEqual, "Doe")
		So(Format("John", "Jane", "Doe"), ShouldEqual, "John Jane Doe")

		So(Format("正人", "", "田中"), ShouldEqual, "田中正人")
		So(Format("正人", "一", "田中"), ShouldEqual, "田中一正人")
		So(Format("ひかる", "", "鈴木"), ShouldEqual, "鈴木ひかる")
		So(Format("Hikaru", "", "Suzuki"), ShouldEqual, "Hikaru Suzuki")

		So(Format("小明", "", "陳"), ShouldEqual, "陳小明")
		So(Format("小明", "大", "陳"), ShouldEqual, "陳大小明")
		So(Format("Siu Ming", "", "Chan"), ShouldEqual, "Siu Ming Chan")
	})
}
