package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileServer(t *testing.T) {
	Convey("FileServer", t, func() {
		Convey("no index.html", func() {
			dir := http.Dir("testdata/noindex")

			h := &FileServer{
				FileSystem:          dir,
				FallbackToIndexHTML: false,
			}

			r, _ := http.NewRequest("GET", "/a-deadbeef.js", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			// Depending on mime.types on the system, the result could be application/javascript or text/javascript.
			So(w.Result().Header.Get("content-type"), ShouldContainSubstring, "javascript")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "public, max-age=604800")

			r, _ = http.NewRequest("GET", "/no-such-file", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")
			So(w.Result().Header.Get("content-type"), ShouldEqual, "")
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")
		})

		Convey("fallback to index.html", func() {
			dir := http.Dir("testdata/index")

			h := &FileServer{
				FileSystem:          dir,
				AssetsDir:           "shared-assets",
				FallbackToIndexHTML: true,
			}

			r, _ := http.NewRequest("GET", "/shared-assets/a-deadbeef.js", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			// Depending on mime.types on the system, the result could be application/javascript or text/javascript.
			So(w.Result().Header.Get("content-type"), ShouldContainSubstring, "javascript")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "public, max-age=604800")

			r, _ = http.NewRequest("GET", "/shared-assets/b-deadbeef.js", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")
			So(w.Result().Header.Get("content-type"), ShouldEqual, "")
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")

			r, _ = http.NewRequest("GET", "/", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("content-type"), ShouldEqual, "text/html; charset=utf-8")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")

			r, _ = http.NewRequest("GET", "/some/route", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("content-type"), ShouldEqual, "text/html; charset=utf-8")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")

			r, _ = http.NewRequest("GET", "/index.html", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("content-type"), ShouldEqual, "text/html; charset=utf-8")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")

			r, _ = http.NewRequest("GET", "/oauth-redirect?code=1234", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("content-type"), ShouldEqual, "text/html; charset=utf-8")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")
		})
	})
}
