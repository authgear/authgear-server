package httputil_test

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestGetIP(t *testing.T) {
	Convey("GetIP", t, func() {
		r, _ := http.NewRequest("POST", "/", nil)
		Convey("should resolve X-Real-IP", func() {
			r.Header.Set("X-Real-IP", "169.254.198.67")
			So(httputil.GetIP(r, true), ShouldEqual, "169.254.198.67")
		})
		Convey("should resolve X-Forwarded-For", func() {
			r.Header.Set("X-Forwarded-For", "[::1]:20595, 169.254.198.67")
			So(httputil.GetIP(r, true), ShouldEqual, "::1")
		})
		Convey("should resolve X-Original-For", func() {
			r.Header.Set("X-Original-For", "[::1]:20595, 169.254.198.67")
			So(httputil.GetIP(r, true), ShouldEqual, "::1")
		})
		Convey("should resolve Forwarded", func() {
			r.Header.Set("Forwarded", "for=127.0.0.1:313;by=169.254.198.67, for=169.254.198.67")
			So(httputil.GetIP(r, true), ShouldEqual, "127.0.0.1")
		})
		Convey("should resolve RemoteAddr", func() {
			r.RemoteAddr = "1.1.1.1:7236"
			So(httputil.GetIP(r, true), ShouldEqual, "1.1.1.1")
		})
		Convey("should resolve with priority", func() {
			r.Header.Set("X-Real-IP", "a")
			r.Header.Set("X-Forwarded-For", "b")
			r.Header.Set("Forwarded", "for=c")
			r.RemoteAddr = "d"
			r.Header.Set("X-Original-For", "e")

			So(httputil.GetIP(r, true), ShouldEqual, "c")

			r.Header.Del("Forwarded")
			So(httputil.GetIP(r, true), ShouldEqual, "b")

			r.Header.Del("X-Forwarded-For")
			So(httputil.GetIP(r, true), ShouldEqual, "e")

			r.Header.Del("X-Original-For")
			So(httputil.GetIP(r, true), ShouldEqual, "a")

			r.Header.Del("X-Real-IP")
			So(httputil.GetIP(r, true), ShouldEqual, "d")
		})
		Convey("should ignore headers when not trusting proxy", func() {
			r.Header.Set("X-Real-IP", "a")
			r.Header.Set("X-Forwarded-For", "b")
			r.Header.Set("Forwarded", "for=c")
			r.RemoteAddr = "d"
			r.Header.Set("X-Original-For", "e")

			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("Forwarded")
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("X-Forwarded-For")
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("X-Original-For")
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("X-Real-IP")
			So(httputil.GetIP(r, false), ShouldEqual, "d")
		})
		Convey("should resolve the browser through the portal's Admin API proxy", func() {
			// A request reaching the Admin API through the portal has passed
			// two proxies, each appending rather than replacing: the edge
			// records the browser, then httputil.ReverseProxy in
			// pkg/portal/transport/admin_api_handler.go appends the address
			// the portal itself saw. The browser must stay the resolved IP, or
			// the Admin API's per-IP rate limits would key every proxied
			// request on the portal's egress and bind all portal users
			// together. See docs/specs/rate-limit.md § Admin API mutations.
			r.RemoteAddr = "10.0.0.9:44321"
			r.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.4, 10.0.0.9")

			So(httputil.GetIP(r, true), ShouldEqual, "203.0.113.7")

			// Without TRUST_PROXY every proxied request collapses onto the
			// portal's address, which over-restricts rather than
			// under-restricts but is not the intended behaviour.
			So(httputil.GetIP(r, false), ShouldEqual, "10.0.0.9")
		})
	})
}
