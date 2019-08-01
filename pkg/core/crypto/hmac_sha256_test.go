package crypto

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHMACSHA256String(t *testing.T) {
	Convey("HMACSHA256", t, func() {
		So(HMACSHA256String([]byte("secret"), []byte("body")), ShouldEqual, "dc46983557fea127b43af721467eb9b3fde2338fe3e14f51952aa8478c13d355")
	})
}
