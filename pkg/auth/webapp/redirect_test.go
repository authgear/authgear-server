package webapp

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRedirectURI(t *testing.T) {
	Convey("GetRedirectURI", t, func() {
		Convey("redirect to default if redirect_uri is absent", func() {
			r, _ := http.NewRequest("GET", "http://example.com", nil)
			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "/settings")
		})

		Convey("redirect to redirect_uri", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/oauth/authorize?client_id=client_id")
		})

		Convey("redirect to redirect_uri when request URI does not have scheme nor host", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Path: "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "/oauth/authorize?client_id=client_id")
		})

		Convey("redirect to redirect_uri with percent encoding", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/oauth/authorize?client_id=client_id&scope=openid+offline_access&ui_locales=en%20zh"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/oauth/authorize?client_id=client_id&scope=openid+offline_access&ui_locales=en%20zh")
		})

		Convey("redirect to relative URI without .", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"relative"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/relative")
		})

		Convey("redirect to relative URI with .", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"./relative"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/relative")
		})

		Convey("redirect to explicit same-origin URI", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"http://example.com/a"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/a")
		})

		Convey("do not redirect to other origin", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"http://evil.com"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "/settings")
		})

		Convey("prevent recursion if redirect_uri is given externally", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
				RawQuery: url.Values{
					"redirect_uri": []string{"/"},
				}.Encode(),
			}).String(), nil)

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "/settings")
		})

		Convey("allow recursion if redirect_uri is given internally", func() {
			r, _ := http.NewRequest("GET", (&url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			}).String(), nil)
			r.Form = url.Values{
				"redirect_uri": []string{"/"},
			}

			So(GetRedirectURI(r, true, "/settings"), ShouldEqual, "http://example.com/")
		})
	})
}
