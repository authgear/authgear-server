package libmagic

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLite(t *testing.T) {
	Convey("Lite", t, func() {
		So(func() {
			MimeFromBytes([]byte{0x23, 0x21})
		}, ShouldNotPanic)
	})
}
