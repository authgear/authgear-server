package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsLikeRollupDefaultAssetName(t *testing.T) {
	Convey("IsLikeRollupDefaultAssetName", t, func() {
		test := func(p string, expected bool) {
			Convey(fmt.Sprintf("%v", p), func() {
				actual := IsLikeRollupDefaultAssetName(p)
				So(actual, ShouldEqual, expected)
			})
		}

		test("", false)
		test("/", false)
		test("/a", false)
		test("/a.js", false)
		test("/a.js.map", false)

		test("/a-deadbeef.js", true)
		test("/a-deadbeef.js.map", true)

		test("/.deadbeef.js", false)
		test("/.deadbeef.js.map", false)

		test("/nested/a-deadbeef.js", true)
		test("/nested/a-deadbeef.js.map", true)

		test("/a-0123456.js", false)
		test("/a-0123456.js.map", false)
	})
}

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
				FallbackToIndexHTML: true,
			}

			r, _ := http.NewRequest("GET", "/a-deadbeef.js", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			// Depending on mime.types on the system, the result could be application/javascript or text/javascript.
			So(w.Result().Header.Get("content-type"), ShouldContainSubstring, "javascript")
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "public, max-age=604800")

			r, _ = http.NewRequest("GET", "/b-deadbeef.js", nil)
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
		})
	})
}
