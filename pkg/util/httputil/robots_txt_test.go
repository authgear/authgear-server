package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var expectedRobotsTxt = `User-agent: *
Disallow: /
`

func TestRobotsTXTHandler(t *testing.T) {
	Convey("RobotsTXTHandler", t, func() {
		w := httptest.NewRecorder()
		h := http.HandlerFunc(RobotsTXTHandler)
		r, _ := http.NewRequest("GET", "", nil)

		h.ServeHTTP(w, r)
		So(w.Result().StatusCode, ShouldEqual, 200)
		So(w.Result().Header.Get("Content-Type"), ShouldEqual, "text/plain")
		So(w.Result().Header.Get("Content-Length"), ShouldEqual, "26")
		So(w.Body.Bytes(), ShouldResemble, []byte(expectedRobotsTxt))
	})
}
