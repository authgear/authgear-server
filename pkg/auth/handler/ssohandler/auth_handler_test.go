package ssohandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	tenantConfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
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
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

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
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.RoleStore = role.NewMockStore()
		sh.AuthHandlerHTMLProvider = sso.NewAuthHandlerHTMLProvider(
			"https://api.example.com",
			"https://api.example.com/skygear.js",
		)

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

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			So(resp.Header().Get("Location"), ShouldEqual, "http://localhost:3000")

			// check cookies
			// it should have following format
			// {
			// 	"callback_url": "callback_url"
			// 	"result": authResp
			// }
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

			// decoded it first
			decoded, err := base64.StdEncoding.DecodeString(ssoDataCookie.Value)
			So(err, ShouldBeNil)
			So(decoded, ShouldNotBeNil)

			// Unmarshal to map
			data := make(map[string]interface{})
			err = json.Unmarshal(decoded, &data)
			So(err, ShouldBeNil)

			// check callback_url
			So(data["callback_url"], ShouldEqual, "http://localhost:3000")

			// check result(authResp)
			authResp, err := json.Marshal(data["result"])
			So(err, ShouldBeNil)
			p, err := sh.OAuthAuthProvider.GetPrincipalByUserID("mock", "mock_user_id")
			So(err, ShouldBeNil)
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(authResp, ShouldEqualJSON, fmt.Sprintf(`{
				"user_id": "%s",
				"profile": {
					"_access": null,
					"_created_at": "0001-01-01T00:00:00Z",
					"_created_by": "",
					"_id": "",
					"_ownerID": "",
					"_recordID": "",
					"_recordType": "",
					"_type": "",
					"_updated_at": "0001-01-01T00:00:00Z",
					"_updated_by": ""
				},
				"access_token": "%s"
			}`, p.UserID, token.AccessToken))
		})

		Convey("should return html page when ux_mode is web_popup", func() {
			action := "login"
			UXMode := "web_popup"

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

			sh.ServeHTTP(resp, req)
			// for web_redirect, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 200)
			JSSKDURLPattern := `<script type="text/javascript" src="https://api.example.com/skygear.js"></script>`
			matched, err := regexp.MatchString(JSSKDURLPattern, resp.Body.String())
			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})

		Convey("should return callback url with result query parameter when ux_mode is ios or android", func() {
			action := "login"
			UXMode := "ios"

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

			sh.ServeHTTP(resp, req)
			// for ios or android, it should redirect to original callback url
			So(resp.Code, ShouldEqual, 302)
			// check location result query parameter
			location, _ := url.Parse(resp.Header().Get("Location"))
			q := location.Query()
			result := q.Get("result")
			decoded, _ := base64.StdEncoding.DecodeString(result)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByUserID("mock", "mock_user_id")
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(decoded, ShouldEqualJSON, fmt.Sprintf(`{
				"user_id": "%s",
				"profile": {
					"_access": null,
					"_created_at": "0001-01-01T00:00:00Z",
					"_created_by": "",
					"_id": "",
					"_ownerID": "",
					"_recordID": "",
					"_recordType": "",
					"_type": "",
					"_updated_at": "0001-01-01T00:00:00Z",
					"_updated_by": ""
				},
				"access_token": "%s"
			}`, p.UserID, token.AccessToken))
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
