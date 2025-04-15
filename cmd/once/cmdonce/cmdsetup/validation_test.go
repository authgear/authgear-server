package cmdsetup

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateEmailAddress(t *testing.T) {
	Convey("validateEmailAddress", t, func() {
		valid := func(input string) {
			err := validateEmailAddress(input)
			So(err, ShouldBeNil)
		}
		invalid := func(input string) {
			err := validateEmailAddress(input)
			So(err, ShouldNotBeNil)
		}

		valid("user@example.com")

		invalid("")
		invalid(" user@example.com")
		invalid("John Doe <user@example.com>")
	})
}

func TestValidateDomain(t *testing.T) {
	Convey("validateDomain", t, func() {
		valid := func(input string) {
			err := validateDomain(input)
			So(err, ShouldBeNil)
		}
		invalid := func(input string) {
			err := validateDomain(input)
			So(err, ShouldNotBeNil)
		}

		valid("localhost")
		valid("example.com")
		valid("a-b.example.com")

		invalid("127.0.0.1")
		invalid("::1")
		invalid("localhost:80")
		invalid("-.example.com")
	})
}

func TestValidatePort(t *testing.T) {
	Convey("validatePort", t, func() {
		valid := func(input string) {
			err := validatePort(input)
			So(err, ShouldBeNil)
		}
		invalid := func(input string) {
			err := validatePort(input)
			So(err, ShouldNotBeNil)
		}

		valid("25")
		valid("587")
		valid("2525")

		invalid("")
		invalid("0")
		invalid("65536")
		invalid("foobar")
		invalid("-25")
		invalid("-587")
		invalid("-2525")
	})
}
