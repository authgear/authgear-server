package sso

import (
	"net/url"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthURLPayload(t *testing.T) {
	Convey("Test AuthURLRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := AuthURLRequestPayload{
				CallbackURL: "callbackURL",
				UXMode:      sso.UXModeWebRedirect,
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without callback url", func() {
			payload := AuthURLRequestPayload{
				UXMode: sso.UXModeWebRedirect,
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without UX mode", func() {
			payload := AuthURLRequestPayload{
				CallbackURL: "callbackURL",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestAuthURLHandler(t *testing.T) {
	Convey("Test TestAuthURLHandler", t, func() {
		h := &AuthURLHandler{}
		h.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		oauthConfig := coreconfig.OAuthConfiguration{
			URLPrefix:      "http://localhost:3000/auth",
			StateJWTSecret: "secret",
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           "mock",
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
			Scope:        "openid profile email",
		}
		mockProvider := sso.MockSSOProvider{
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
		}
		h.Provider = &mockProvider
		h.Action = "login"
		payload := AuthURLRequestPayload{
			CallbackURL: "callbackURL",
			UXMode:      sso.UXModeWebRedirect,
			Options: map[string]interface{}{
				"number": 1,
			},
		}

		Convey("should return login_auth_url", func() {
			resp, err := h.Handle(payload)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			// check base url
			u, _ := url.Parse(resp.(string))
			So(u.Host, ShouldEqual, "mock")
			So(u.Path, ShouldEqual, "/auth")

			// check querys
			q := u.Query()
			So(q.Get("response_type"), ShouldEqual, "code")
			So(q.Get("client_id"), ShouldEqual, "mock_client_id")
			So(q.Get("scope"), ShouldEqual, "openid profile email")
			So(q.Get("number"), ShouldEqual, "1")

			// check redirect_uri
			r, _ := url.Parse(q.Get("redirect_uri"))
			So(r.Host, ShouldEqual, "localhost:3000")
			So(r.Path, ShouldEqual, "/auth/sso/mock/auth_handler")

			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			_, err = jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			})
			So(err, ShouldBeNil)
			So(claims.State.UXMode, ShouldEqual, sso.UXModeWebRedirect)
			So(claims.State.CallbackURL, ShouldEqual, "callbackURL")
			So(claims.State.Action, ShouldEqual, "login")
			So(claims.State.UserID, ShouldEqual, "faseng.cat.id")
		})

		Convey("should return link_auth_url", func() {
			h.Action = "link"
			resp, err := h.Handle(payload)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			u, _ := url.Parse(resp.(string))
			q := u.Query()
			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			})
			So(claims.State.Action, ShouldEqual, "link")
		})
	})
}
