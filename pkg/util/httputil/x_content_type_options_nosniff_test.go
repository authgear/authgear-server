package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestXContentTypeOptionsNosniff(t *testing.T) {
	Convey("XContentTypeOptionsNosniff", t, func() {
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := XContentTypeOptionsNosniff(dummy)
			return h
		}

		Convey("disable content type sniffing", func() {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
		})
	})
}
