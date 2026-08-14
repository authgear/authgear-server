package sentry_test

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/sentry"
)

func TestMakeMinimalRequest(t *testing.T) {
	Convey("MakeMinimalRequest", t, func() {
		Convey("should make a request mirroring the incoming request", func() {
			r, _ := http.NewRequest("GET", "/path", nil)
			r.Host = "example.com"

			req := sentry.MakeMinimalRequest(r, false)
			So(req.URL.String(), ShouldEqual, "http://example.com/path")
		})

		Convey("should panic with the underlying error instead of returning a nil request", func() {
			r, _ := http.NewRequest("GET", "/path", nil)
			r.Header.Set("X-Forwarded-Host", "example.com evil.com")

			So(func() { sentry.MakeMinimalRequest(r, true) }, ShouldPanic)
		})
	})
}
