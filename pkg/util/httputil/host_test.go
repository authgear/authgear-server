package httputil_test

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestGetHost(t *testing.T) {
	Convey("GetHost", t, func() {
		r, _ := http.NewRequest("POST", "/", nil)

		Convey("should resolve X-Forwarded-Host", func() {
			r.Header.Set("X-Forwarded-Host", "example.com")
			So(httputil.GetHost(r, true), ShouldEqual, "example.com")
		})

		Convey("should resolve X-Original-Host", func() {
			r.Header.Set("X-Original-Host", "example.com")
			So(httputil.GetHost(r, true), ShouldEqual, "example.com")
		})

		Convey("should resolve with priority", func() {
			r.Header.Set("X-Forwarded-Host", "a")
			r.Header.Set("X-Original-Host", "b")

			So(httputil.GetHost(r, true), ShouldEqual, "a")

			r.Header.Del("X-Forwarded-Host")
			So(httputil.GetHost(r, true), ShouldEqual, "b")

			r.Header.Del("X-Original-Host")
			So(httputil.GetHost(r, true), ShouldEqual, "")
		})

		Convey("should ignore headers when not trusting proxy", func() {
			r.Header.Set("X-Forwarded-Host", "a")
			r.Header.Set("X-Original-Host", "b")

			So(httputil.GetHost(r, false), ShouldEqual, "")

			r.Header.Del("X-Forwarded-Host")
			So(httputil.GetHost(r, false), ShouldEqual, "")

			r.Header.Del("X-Original-Host")
			So(httputil.GetHost(r, false), ShouldEqual, "")
		})
	})
}
