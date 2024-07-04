package viewmodels

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestComposeAuthUIWindowMessageAllowedOrigin(t *testing.T) {
	Convey("composeAuthUIWindowMessageAllowedOrigin", t, func() {
		Convey("Given a origin", func() {
			origin := "http://www.example.com"
			Convey("It returns the origin unprocessed", func() {
				requestProto := "http"
				So(composeAuthUIWindowMessageAllowedOrigin(origin, requestProto), ShouldEqual, origin)
			})
		})

		Convey("Given a host", func() {
			host := "www.example.com"
			Convey("It returns the origin according to assgined proto", func() {
				requestProto := "http"
				So(composeAuthUIWindowMessageAllowedOrigin(host, requestProto), ShouldEqual, "http://www.example.com")
			})
		})
	})
}
