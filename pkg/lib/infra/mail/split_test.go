package mail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSplitAddress(t *testing.T) {
	test := func(addr string, local string, domain string) {
		l, d := SplitAddress(addr)
		So(l, ShouldEqual, local)
		So(d, ShouldEqual, domain)
	}

	Convey("SplitAdress", t, func() {
		test("user@example.com", "user", "example.com")
		test("johndoe@example.com", "johndoe", "example.com")
	})
}
