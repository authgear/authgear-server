package checksum

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCRC32IEEEInHex(t *testing.T) {
	Convey("CRC32IEEEInHex", t, func() {
		So(CRC32IEEEInHex([]byte("data")), ShouldEqual, "00000000adf3f363")
	})
}
