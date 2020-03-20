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
			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
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

			RedirectToRedirectURI(w, r)
			check(w, DefaultRedirectURI)
		})

		Convey("prevent recursion", func() {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/"},
				}.Encode(),
			}).String(), nil)

			RedirectToRedirectURI(w, r)
			check(w, DefaultRedirectURI)
		})
	})
}
