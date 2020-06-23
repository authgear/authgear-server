package webapp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRedirectToRedirectURI(t *testing.T) {
	Convey("RedirectToRedirectURI", t, func() {
		check := func(w *httptest.ResponseRecorder, redirectURI string) {
			So(w.Result().StatusCode, ShouldEqual, http.StatusFound)
			So(w.Result().Header.Get("Location"), ShouldEqual, redirectURI)
		}

		Convey("redirect to default if redirect_uri is absent", func() {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "http://example.com", nil)
			RedirectToRedirectURI(w, r, true)
			check(w, DefaultRedirectURI)
		})

		Convey("redirect to redirect_uri", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/oauth/authorize?client_id=client_id")
		})

		Convey("redirect to redirect_uri when request URI does not have scheme nor host", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Path: "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "/oauth/authorize?client_id=client_id")
		})

		Convey("redirect to redirect_uri with percent encoding", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id&scope=openid+offline_access&ui_locales=en%20zh"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/oauth/authorize?client_id=client_id&scope=openid+offline_access&ui_locales=en%20zh")
		})

		Convey("redirect to relative URI without .", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"relative"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/relative")
		})

		Convey("redirect to relative URI with .", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"./relative"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/relative")
		})

		Convey("redirect to explicit same-origin URI", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"http://example.com/a"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/a")
		})

		Convey("do not redirect to other origin", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"http://evil.com"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, DefaultRedirectURI)
		})

		Convey("prevent recursion if redirect_uri is given externally", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r, true)
			check(w, DefaultRedirectURI)
		})

		Convey("allow recursion if redirect_uri is given internally", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			}).String(), nil)
			r.Form = url.Values{
				"redirect_uri": []string{"/"},
			}

			RedirectToRedirectURI(w, r, true)
			check(w, "http://example.com/")
		})
	})
}

func TestMakeURLWithPathWithX(t *testing.T) {
	Convey("MakeURLWithPathWithX", t, func() {
		test := func(str string, path string, expected string) {
			u, err := url.Parse(str)
			So(err, ShouldBeNil)
			actual := MakeURLWithPathWithX(u, path)
			So(actual, ShouldEqual, expected)
		}

		test("http://example.com", "/login", "/login")
		test("http://example.com?a=a", "/login", "/login?a=a")
		test("http://example.com/login?a=a", "/login", "/login?a=a")
		test("http://example.com/login?a=a&x_a=a", "/login", "/login?a=a&x_a=a")
	})
}

func TestMakeURLWithPathWithoutX(t *testing.T) {
	Convey("MakeURLWithPathWithoutX", t, func() {
		test := func(str string, path string, expected string) {
			u, err := url.Parse(str)
			So(err, ShouldBeNil)
			actual := MakeURLWithPathWithoutX(u, path)
			So(actual, ShouldEqual, expected)
		}

		test("http://example.com", "/login", "/login")
		test("http://example.com?a=a", "/login", "/login?a=a")
		test("http://example.com/login?a=a", "/login", "/login?a=a")
		test("http://example.com/login?a=a&x_a=a", "/login", "/login?a=a")
	})
}

func TestMakeURLWithQuery(t *testing.T) {
	Convey("MakeURLWithQuery", t, func() {
		test := func(str string, name string, value string, expected string) {
			u, err := url.Parse(str)
			So(err, ShouldBeNil)
			actual := MakeURLWithQuery(u, url.Values{
				name: []string{value},
			})
			So(actual, ShouldEqual, expected)
		}

		test("http://example.com", "a", "b", "?a=b")
		test("http://example.com?c=d", "a", "b", "?a=b&c=d")
	})
}

func TestNewURLWithPathAndQuery(t *testing.T) {
	Convey("NewURLWithPathAndQuery", t, func() {
		test := func(path string, name string, value string, expected string) {
			actual := NewURLWithPathAndQuery(path, url.Values{
				name: []string{value},
			})
			So(actual, ShouldEqual, expected)
		}

		test("/login", "a", "b c", "/login?a=b+c")
		test("/", "a", "b", "/?a=b")
	})
}
