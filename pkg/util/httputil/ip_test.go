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
			So(httputil.GetIP(r, true), ShouldEqual, "c")

			r.Header.Del("Forwarded")
			So(httputil.GetIP(r, true), ShouldEqual, "b")

			r.Header.Del("X-Forwarded-For")
			So(httputil.GetIP(r, true), ShouldEqual, "a")

			r.Header.Del("X-Real-IP")
			So(httputil.GetIP(r, true), ShouldEqual, "d")
		})
		Convey("should ignore headers when not trusting proxy", func() {
			r.Header.Set("X-Real-IP", "a")
			r.Header.Set("X-Forwarded-For", "b")
			r.Header.Set("Forwarded", "for=c")
			r.RemoteAddr = "d"
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("Forwarded")
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("X-Forwarded-For")
			So(httputil.GetIP(r, false), ShouldEqual, "d")

			r.Header.Del("X-Real-IP")
			So(httputil.GetIP(r, false), ShouldEqual, "d")
		})
	})
}
