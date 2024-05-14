package handler_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
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

	Convey("ProxyRedirectHandler", t, func() {
		Convey("should allow allowlisted", func() {
			var err error

			err = h.Validate("http://app.example.com/auth?query=test#abc")
			So(err, ShouldBeNil)

			err = h.Validate("com.example.myapp://host/path?code=code")
			So(err, ShouldBeNil)

			err = h.Validate("com.example.myapp://host/path?error=cancel")
			So(err, ShouldBeNil)
		})

		Convey("should allow all path under custom ui", func() {
			var err error

			err = h.Validate("http://authui.example.com/auth/complete?state=state")
			So(err, ShouldBeNil)

			err = h.Validate("http://authui.example.com/error?error=")
			So(err, ShouldBeNil)
		})

		Convey("should reject uri that not on the list", func() {
			var err error

			err = h.Validate("http://app2.example.com/auth/complete?state=state")
			So(err, ShouldBeError, "redirect URI is not allowed")
		})
	})

}
