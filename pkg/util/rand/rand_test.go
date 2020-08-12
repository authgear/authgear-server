package rand

import (
	mrand "math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRandStringWithCharset(t *testing.T) {
	Convey("RandStringWithCharset", t, func() {
		alphabet := "0123456789"
		length := 10
		rands := []*mrand.Rand{SecureRand, InsecureRand}
		for _, r := range rands {
			for i := 0; i < 1000; i++ {
				out := StringWithAlphabet(length, alphabet, r)
				for _, run := range out {
					So(run >= '0', ShouldBeTrue)
					So(run <= '9', ShouldBeTrue)
				}
				So(len(out), ShouldEqual, length)
			}
		}
	})
}
