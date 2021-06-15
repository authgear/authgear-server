package httputil_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestGetProto(t *testing.T) {
	Convey("GetProto", t, func() {
		r, _ := http.NewRequest("POST", "/", nil)

		Convey("should resolve X-Forwarded-Proto", func() {
			r.Header.Set("X-Forwarded-Proto", "https")
			So(httputil.GetProto(r, true), ShouldEqual, "https")
		})

		Convey("should resolve X-Original-Proto", func() {
			r.Header.Set("X-Original-Proto", "https")
			So(httputil.GetProto(r, true), ShouldEqual, "https")
		})

		Convey("should resolve http.Request.TLS", func() {
			r.TLS = &tls.ConnectionState{}
			So(httputil.GetProto(r, true), ShouldEqual, "https")
		})

		Convey("should resolve to plain HTTP", func() {
			So(httputil.GetProto(r, true), ShouldEqual, "http")
		})

		Convey("should resolve with priority", func() {
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("X-Original-Proto", "https")
			r.TLS = &tls.ConnectionState{}

			So(httputil.GetProto(r, true), ShouldEqual, "https")

			r.Header.Del("X-Forwarded-Proto")
			So(httputil.GetProto(r, true), ShouldEqual, "https")

			r.Header.Del("X-Original-Proto")
			So(httputil.GetProto(r, true), ShouldEqual, "https")

			r.TLS = nil
			So(httputil.GetProto(r, true), ShouldEqual, "http")
		})
	})
}
