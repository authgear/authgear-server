package handler_test

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
)

func TestProxyRedirectHandler(t *testing.T) {
	h := handler.ProxyRedirectHandler{
		OAuthConfig: &config.OAuthConfig{
			Clients: []config.OAuthClientConfig{
				{
					RedirectURIs: []string{"http://app.example.com/auth"},
					CustomUIURI:  "http://authui.example.com",
				},
				{
					RedirectURIs: []string{"com.example.myapp://host/path"},
				},
			},
		},
		HTTPOrigin: "http://auth.example.com",
	}

	testError := func(input string, error string) {
		_, err := h.Validate(input)
		So(err, ShouldNotBeNil)
		So(err, ShouldBeError, error)
	}

	testOK := func(input string, expected *oauth.WriteResponseOptions) {
		actual, err := h.Validate(input)
		So(err, ShouldBeNil)
		So(actual, ShouldEqual, expected)
	}

	Convey("ProxyRedirectHandler", t, func() {
		Convey("should allow allowlisted", func() {
			testOK("http://app.example.com/auth?query=test#abc", &oauth.WriteResponseOptions{
				RedirectURI: &url.URL{
					Scheme:   "http",
					Host:     "app.example.com",
					Path:     "/auth",
					RawQuery: "query=test",
					Fragment: "abc",
				},
				ResponseMode: "query",
				UseHTTP200:   true,
				Response:     make(map[string]string),
			})

			testOK("com.example.myapp://host/path?code=code", &oauth.WriteResponseOptions{
				RedirectURI: &url.URL{
					Scheme:   "com.example.myapp",
					Host:     "host",
					Path:     "/path",
					RawQuery: "code=code",
				},
				ResponseMode: "query",
				UseHTTP200:   false,
				Response:     make(map[string]string),
			})

			testOK("com.example.myapp://host/path?error=cancel", &oauth.WriteResponseOptions{
				RedirectURI: &url.URL{
					Scheme:   "com.example.myapp",
					Host:     "host",
					Path:     "/path",
					RawQuery: "error=cancel",
				},
				ResponseMode: "query",
				UseHTTP200:   false,
				Response:     make(map[string]string),
			})
		})

		Convey("should allow all path under custom ui", func() {
			testOK("http://authui.example.com/auth/complete?state=state", &oauth.WriteResponseOptions{
				RedirectURI: &url.URL{
					Scheme:   "http",
					Host:     "authui.example.com",
					Path:     "/auth/complete",
					RawQuery: "state=state",
				},
				ResponseMode: "query",
				UseHTTP200:   true,
				Response:     make(map[string]string),
			})

			testOK("http://authui.example.com/error?error=", &oauth.WriteResponseOptions{
				RedirectURI: &url.URL{
					Scheme:   "http",
					Host:     "authui.example.com",
					Path:     "/error",
					RawQuery: "error=",
				},
				ResponseMode: "query",
				UseHTTP200:   true,
				Response:     make(map[string]string),
			})
		})

		Convey("should reject uri that not on the list", func() {
			testError("http://app2.example.com/auth/complete?state=state", "redirect URI is not allowed")
		})
	})

}
