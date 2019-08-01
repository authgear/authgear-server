package crypto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSHA256String(t *testing.T) {
	Convey("SHA256String", t, func() {
		So(SHA256String("hello world\n"), ShouldEqual, "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447")
	})
}
