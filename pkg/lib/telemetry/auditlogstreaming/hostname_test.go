package auditlogstreaming

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResolveHostname(t *testing.T) {
	Convey("ResolveHostname", t, func() {
		Convey("returns the bare host", func() {
			So(ResolveHostname("https://myproject.authgear.cloud"), ShouldEqual, "myproject.authgear.cloud")
		})

		Convey("strips the port -- the case that matters for local development and e2e", func() {
			So(ResolveHostname("http://app.localhost:4000"), ShouldEqual, "app.localhost")
		})

		Convey("returns empty for a host containing non-ASCII", func() {
			So(ResolveHostname("https://éxample.com"), ShouldEqual, "")
		})

		Convey("returns empty for an unparseable origin", func() {
			So(ResolveHostname("http://[::1"), ShouldEqual, "")
		})
	})
}
