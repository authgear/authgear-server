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

		Convey("source map is not served by default", func() {
			dir := http.Dir("testdata/noindex")

			h := &FileServer{
				FileSystem:          dir,
				FallbackToIndexHTML: false,
			}

			r, _ := http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-cache")
			So(w.Result().Header.Get("content-type"), ShouldEqual, "")
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")

			// The extension is matched case-insensitively, so that the check
			// stays correct on a case-insensitive file system.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js.MAP", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")

			// A trailing slash is normalized away before the check.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js.map/", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")
		})

		Convey("source map is not served by default, even when index.html is the fallback", func() {
			dir := http.Dir("testdata/index")

			h := &FileServer{
				FileSystem:          dir,
				AssetsDir:           "shared-assets",
				FallbackToIndexHTML: true,
			}

			// A source map that exists.
			r, _ := http.NewRequest("GET", "/shared-assets/a-deadbeef.js.map", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")

			// A source map that does not exist MUST NOT fall back to index.html.
			r, _ = http.NewRequest("GET", "/no-such-file.js.map", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")
		})

		Convey("source map is served when it is enabled", func() {
			dir := http.Dir("testdata/noindex")

			h := &FileServer{
				FileSystem:          dir,
				FallbackToIndexHTML: false,
				SourceMap: SourceMapConfig{
					Enabled: true,
				},
			}

			r, _ := http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Body.String(), ShouldContainSubstring, "sourcesContent")

			// Non source map files are unaffected.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("source map served from the assets dir is never publicly cacheable", func() {
			dir := http.Dir("testdata/index")

			h := &FileServer{
				FileSystem:          dir,
				AssetsDir:           "shared-assets",
				FallbackToIndexHTML: true,
				SourceMap: SourceMapConfig{
					Enabled: true,
				},
			}

			// A source map can be protected by credentials,
			// so it must not be stored by a shared cache.
			r, _ := http.NewRequest("GET", "/shared-assets/a-deadbeef.js.map", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "no-store")

			// Other name-hashed assets keep the long-lived public cache.
			r, _ = http.NewRequest("GET", "/shared-assets/a-deadbeef.js", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Result().Header.Get("cache-control"), ShouldEqual, "public, max-age=604800")
		})

		Convey("source map is protected by basic auth when a sentry token is set", func() {
			dir := http.Dir("testdata/noindex")

			h := &FileServer{
				FileSystem:          dir,
				FallbackToIndexHTML: false,
				SourceMap: SourceMapConfig{
					Enabled:     true,
					SentryToken: "s3cret",
				},
			}

			// No credentials.
			r, _ := http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 401)
			So(w.Result().Header.Get("www-authenticate"), ShouldEqual, `Basic realm="source map"`)
			So(w.Result().Header.Get("content-type"), ShouldEqual, "")
			So(w.Result().Header.Get("content-length"), ShouldEqual, "0")

			// Wrong password.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			r.SetBasicAuth("sentry", "wrong")
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 401)

			// Correct password. The username is ignored.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			r.SetBasicAuth("", "s3cret")
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
			So(w.Body.String(), ShouldContainSubstring, "sourcesContent")

			// Non source map files are NOT protected.
			r, _ = http.NewRequest("GET", "/a-deadbeef.js", nil)
			w = httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 200)
		})

		Convey("the sentry token does not enable source map serving on its own", func() {
			dir := http.Dir("testdata/noindex")

			h := &FileServer{
				FileSystem:          dir,
				FallbackToIndexHTML: false,
				SourceMap: SourceMapConfig{
					Enabled:     false,
					SentryToken: "s3cret",
				},
			}

			r, _ := http.NewRequest("GET", "/a-deadbeef.js.map", nil)
			r.SetBasicAuth("", "s3cret")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, 404)
		})
	})
}
