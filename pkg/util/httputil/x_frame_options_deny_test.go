package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestXFrameOptionsDeny(t *testing.T) {
	Convey("XFrameOptionsDeny", t, func() {
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := XFrameOptionsDeny(dummy)
			return h
		}

		Convey("output x-frame-options: DENY", func() {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("X-Frame-Options"), ShouldEqual, "DENY")
		})
	})
}
