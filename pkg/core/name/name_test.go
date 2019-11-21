package name

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateAppName(t *testing.T) {
	Convey("ValidateAppName", t, func() {
		test := func(name string, ok bool) {
			So(ValidateAppName(name) == nil, ShouldEqual, ok)
		}
		test("", false)                                          // at least 1 alphanumeric char
		test("-a", false)                                        // cannot start/end with dash
		test("a-", false)                                        // cannot start/end with dash
		test("a", true)                                          // good
		test("123app-production", true)                          // good
		test("a234567890123456789012345678901234567890", true)   // longest possible
		test("a2345678901234567890123456789012345678901", false) // too long
	})
}
