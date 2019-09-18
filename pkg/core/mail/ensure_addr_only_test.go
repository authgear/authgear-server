package mail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEnsureAddressOnly(t *testing.T) {
	Convey("EnsureAddressOnly", t, func() {
		So(EnsureAddressOnly("user@example.com"), ShouldBeNil)
		So(EnsureAddressOnly("User <user@example.com>"), ShouldBeError, "address must not have name")
		So(EnsureAddressOnly(" user@example.com "), ShouldBeError, "formatted address is not the same as input")
		So(EnsureAddressOnly("nonsense"), ShouldNotBeNil)
	})
}
