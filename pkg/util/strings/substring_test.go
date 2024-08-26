package strings

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSubstring(t *testing.T) {
	Convey("GetFirstN", t, func() {
		original := "abcdefghijklmnopqrstuvwxyz"

		negative := func() { GetFirstN(original, -999) }
		empty := GetFirstN(original, 0)
		untilJ := GetFirstN(original, 10)
		untilZ := GetFirstN(original, 26)
		overflow := GetFirstN(original, 9999)

		So(negative, ShouldPanicWith, "n must be >= 0")
		So(empty, ShouldResemble, "")
		So(untilJ, ShouldResemble, "abcdefghij")
		So(untilZ, ShouldResemble, "abcdefghijklmnopqrstuvwxyz")
		So(overflow, ShouldResemble, "abcdefghijklmnopqrstuvwxyz")
	})
}
