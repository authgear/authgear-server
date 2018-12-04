package ssohandler

import (
	"net/url"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthURLPayload(t *testing.T) {
	Convey("Test AuthURLRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := AuthURLRequestPayload{
				CallbackURL: "callbackURL",
				UXMode:      sso.WebRedirect,
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without callback url", func() {
			payload := AuthURLRequestPayload{
				UXMode: sso.WebRedirect,
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

func TestSSOUXModeConvertor(t *testing.T) {
	Convey("Test UXModeFromString", t, func() {
		Convey("should convert string to sso.UXMode", func() {
			modes := []sso.UXMode{
				sso.WebRedirect,
				sso.WebPopup,
				sso.IOS,
				sso.Android,
			}
			for _, m := range modes {
				modeFromStr := sso.UXModeFromString(m.String())
				So(modeFromStr, ShouldEqual, m)
			}
		})

		Convey("should convert unknown string to sso.Undefined", func() {
			modeFromStr := sso.UXModeFromString("some_string")
			So(modeFromStr, ShouldEqual, sso.Undefined)
		})
	})
}

func TestAuthURLHandler(t *testing.T) {
	Convey("Test TestAuthURLHandler", t, func() {
		h := &AuthURLHandler{}
		h.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		setting := sso.Setting{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: "secret",
		}
		config := sso.Config{
			Name:         "mock",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProverImpl{
			BaseURL: "http://mock/auth",
			Setting: setting,
			Config:  config,
		}
		h.Provider = &mockProvider
		h.Action = "login"
		payload := AuthURLRequestPayload{
			Scope: []string{
				"openid",
				"profile",
				"email",
			},
			CallbackURL: "callbackURL",
			UXMode:      sso.WebRedirect,
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
			So(r.Path, ShouldEqual, "/sso/mock/auth_handler")

			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			_, err = jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			})
			So(err, ShouldBeNil)
			So(claims.State.UXMode, ShouldEqual, sso.WebRedirect.String())
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

		Convey("should use default scope when param scope is empty", func() {
			config.Scope = []string{
				"huasheng",
				"zhima",
			}
			mockProvider.Config = config
			// remove payload scope
			payload.Scope = []string{}

			resp, err := h.Handle(payload)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			// check querys
			u, _ := url.Parse(resp.(string))
			q := u.Query()
			So(q.Get("scope"), ShouldEqual, "huasheng zhima")
		})
	})
}
