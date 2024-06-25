package httproute

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute(t *testing.T) {
	Convey("Route methods are immutable", t, func() {
		r1 := Route{}
		r2 := r1.WithMethods("GET")
		So(r1.Methods, ShouldBeNil)
		So(r2.Methods, ShouldResemble, []string{"GET"})
	})

	Convey("Prepend path pattern", t, func() {
		Convey("Prepend path has / suffix and original path has / prefix", func() {
			r1 := Route{
				PathPattern: "/path",
			}
			r2 := r1.PrependPathPattern("/prepend/")
			So(r2.PathPattern, ShouldEqual, "/prepend/path")
		})
		Convey("Prepend path has / suffix and original path has no / prefix", func() {
			r1 := Route{
				PathPattern: "path",
			}
			r2 := r1.PrependPathPattern("/prepend/")
			So(r2.PathPattern, ShouldEqual, "/prepend/path")
		})
		Convey("Prepend path has no / suffix and original path has / prefix", func() {
			r1 := Route{
				PathPattern: "/path",
			}
			r2 := r1.PrependPathPattern("/prepend")
			So(r2.PathPattern, ShouldEqual, "/prepend/path")
		})
		Convey("Prepend path has no / suffix and original path has no / prefix", func() {
			r1 := Route{
				PathPattern: "path",
			}
			r2 := r1.PrependPathPattern("/prepend")
			So(r2.PathPattern, ShouldEqual, "/prepend/path")
		})
	})
}

func TestRedirectTrailingSlash(t *testing.T) {
	Convey("RedirectTrailingSlash", t, func() {
		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		router.Add(Route{
			Methods:     []string{"GET"},
			PathPattern: "/foo",
		}, h)

		r, _ := http.NewRequest("GET", "/foo/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(w.Result().StatusCode, ShouldEqual, 301)
		So(w.Header().Get("Location"), ShouldEqual, "/foo")
	})
}

func TestRedirectFixedPath(t *testing.T) {
	Convey("RedirectFixedPath", t, func() {
		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		router.Add(Route{
			Methods:     []string{"GET"},
			PathPattern: "/foo",
		}, h)

		r, _ := http.NewRequest("GET", "/./FOO", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(w.Result().StatusCode, ShouldEqual, 301)
		So(w.Header().Get("Location"), ShouldEqual, "/foo")
	})
}

func TestHandleMethodNotAllowed(t *testing.T) {
	Convey("HandleMethodNotAllowed", t, func() {
		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
		router.Add(Route{
			Methods:     []string{"GET"},
			PathPattern: "/foo",
		}, h)

		r, _ := http.NewRequest("POST", "/foo", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(w.Result().StatusCode, ShouldEqual, 405)
	})
}

func TestHandleOptions(t *testing.T) {
	Convey("OPTIONS are not handled", t, func() {
		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("handled by handler"))
		})

		router.Add(Route{
			Methods:     []string{"OPTIONS", "GET"},
			PathPattern: "/foo",
		}, h)

		r, _ := http.NewRequest("GET", "/foo", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(w.Body.String(), ShouldEqual, "handled by handler")
	})
}

func TestGetParam(t *testing.T) {
	Convey("GetParam does not crash", t, func() {
		value := "unset"

		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value = GetParam(r, "foobar")
		})

		router.Add(Route{
			Methods:     []string{"GET"},
			PathPattern: "/foo",
		}, h)

		r, _ := http.NewRequest("GET", "/foo", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(value, ShouldEqual, "")
	})

	Convey("GetParam does its job", t, func() {
		var value string

		router := NewRouter()
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value = GetParam(r, "name")
		})

		router.Add(Route{
			Methods:     []string{"GET"},
			PathPattern: "/foo/:name",
		}, h)

		r, _ := http.NewRequest("GET", "/foo/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, r)

		So(value, ShouldEqual, "johndoe")
	})
}

func TestMiddleware(t *testing.T) {
	Convey("Middleware", t, func() {
		var observedLabels []string

		makeMiddleware := func(label string) Middleware {
			return MiddlewareFunc(func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					observedLabels = append(observedLabels, fmt.Sprintf("before %v", label))
					h.ServeHTTP(w, r)
					observedLabels = append(observedLabels, fmt.Sprintf("after %v", label))
				})
			})
		}

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			observedLabels = append(observedLabels, "handler")
		})

		Convey("Single middleware", func() {
			router := NewRouter()
			router.Add(Route{
				Methods:     []string{"GET"},
				PathPattern: "/foo",
				Middleware:  Chain(makeMiddleware("m1")),
			}, h)

			r, _ := http.NewRequest("GET", "/foo", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			So(observedLabels, ShouldResemble, []string{"before m1", "handler", "after m1"})
		})

		Convey("Many middlewares", func() {
			router := NewRouter()
			router.Add(Route{
				Methods:     []string{"GET"},
				PathPattern: "/foo",
				Middleware:  Chain(makeMiddleware("m1"), makeMiddleware("m2")),
			}, h)

			r, _ := http.NewRequest("GET", "/foo", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			So(observedLabels, ShouldResemble, []string{"before m1", "before m2", "handler", "after m2", "after m1"})
		})

		Convey("Nested chain", func() {
			router := NewRouter()
			router.Add(Route{
				Methods:     []string{"GET"},
				PathPattern: "/foo",
				Middleware:  Chain(Chain(makeMiddleware("m1"), makeMiddleware("m2")), Chain(makeMiddleware("m3"), makeMiddleware("m4"))),
			}, h)

			r, _ := http.NewRequest("GET", "/foo", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			So(observedLabels, ShouldResemble, []string{
				"before m1",
				"before m2",
				"before m3",
				"before m4",
				"handler",
				"after m4",
				"after m3",
				"after m2",
				"after m1",
			})
		})
	})
}
