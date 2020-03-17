package authn

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccessEventConnInfoIP(t *testing.T) {
	Convey("AccessEventConnInfo.IP", t, func() {
		Convey("should resolve X-Real-IP", func() {
			ip := AccessEventConnInfo{
				XRealIP: "169.254.198.67",
			}.IP()
			So(ip, ShouldEqual, "169.254.198.67")
		})
		Convey("should resolve X-Forwarded-For", func() {
			ip := AccessEventConnInfo{
				XForwardedFor: "[::1]:20595, 169.254.198.67",
			}.IP()
			So(ip, ShouldEqual, "::1")
		})
		Convey("should resolve Forwarded", func() {
			ip := AccessEventConnInfo{
				Forwarded: "for=127.0.0.1:313;by=169.254.198.67, for=169.254.198.67",
			}.IP()
			So(ip, ShouldEqual, "127.0.0.1")
		})
		Convey("should resolve RemoteAddr", func() {
			ip := AccessEventConnInfo{
				RemoteAddr: "1.1.1.1:7236",
			}.IP()
			So(ip, ShouldEqual, "1.1.1.1")
		})
		Convey("should resolve with priority", func() {
			ip := AccessEventConnInfo{
				XRealIP:       "a",
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			}.IP()
			So(ip, ShouldEqual, "a")

			ip = AccessEventConnInfo{
				XForwardedFor: "b",
				Forwarded:     "for=c",
				RemoteAddr:    "d",
			}.IP()
			So(ip, ShouldEqual, "b")

			ip = AccessEventConnInfo{
				Forwarded:  "for=c",
				RemoteAddr: "d",
			}.IP()
			So(ip, ShouldEqual, "c")

			ip = AccessEventConnInfo{
				RemoteAddr: "d",
			}.IP()
			So(ip, ShouldEqual, "d")
		})
	})
}
