package ssohandler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	tenantConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthPayload(t *testing.T) {
	Convey("Test AuthRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := AuthRequestPayload{
				Code:         "code",
				EncodedState: "state",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without code", func() {
			payload := AuthRequestPayload{
				EncodedState: "state",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without state", func() {
			payload := AuthRequestPayload{
				Code: "code",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestAuthHandler(t *testing.T) {
	Convey("Test TestAuthURLHandler", t, func() {
		stateJWTSecret := "secret"
		sh := &AuthHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		setting := sso.Setting{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
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
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]oauth.Principal{},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewJWTStore("myApp", "secret", 0)
		sh.TokenStore = mockTokenStore
		sh.RoleStore = role.NewMockStore()
		h := sh.Handler()

		// tenant config
		tConfig := tenantConfig.NewTenantConfiguration()
		tConfig.SSOSetting = tenantConfig.SSOSetting{
			URLPrefix:      "http://localhost:3000",
			StateJWTSecret: stateJWTSecret,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}

		Convey("should return callback url when ux_mode is web_redirect and action is login", func() {
			action := "login"
			UXMode := "web_redirect"

			// oauth state
			state := sso.State{
				CallbackURL: "http://localhost:3000",
				UXMode:      UXMode,
				Action:      action,
			}
			encodedState, _ := sso.EncodeState(stateJWTSecret, state)

			v := url.Values{}
			v.Set("code", "code")
			v.Add("state", encodedState)
			u := url.URL{
				RawQuery: v.Encode(),
			}

			req, _ := http.NewRequest("GET", u.RequestURI(), nil)
			tenantConfig.SetTenantConfig(req, tConfig)
			resp := httptest.NewRecorder()

			h.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			cookies := resp.Result().Cookies()
			So(cookies, ShouldNotBeEmpty)
			var ssoDataCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "sso_data" {
					ssoDataCookie = c
					break
				}
			}
			So(ssoDataCookie, ShouldNotBeNil)
			decoded, err := base64.StdEncoding.DecodeString(ssoDataCookie.Value)
			So(err, ShouldBeNil)
			So(decoded, ShouldNotBeNil)
			data := make(map[string]interface{})
			err = json.Unmarshal(decoded, &data)
			So(err, ShouldBeNil)
			So(data["callback_url"], ShouldEqual, "http://localhost:3000")
			// TODO: check authResp by ShouldEqualJSON
			// So(data["result"], ShouldEqualJSON, `{
			// }`)
		})
	})
}

func TestValidateCallbackURL(t *testing.T) {
	Convey("Test ValidateCallbackURL", t, func() {
		sh := &AuthHandler{}
		callbackURL := "http://localhost:3000"
		allowedCallbackURLs := []string{
			"http://localhost",
			"http://127.0.0.1",
		}

		e := sh.validateCallbackURL(allowedCallbackURLs, callbackURL)
		So(e, ShouldBeNil)

		callbackURL = "http://oursky"
		e = sh.validateCallbackURL(allowedCallbackURLs, callbackURL)
		So(e, ShouldNotBeNil)

		e = sh.validateCallbackURL(allowedCallbackURLs, "")
		So(e, ShouldNotBeNil)
	})
}
