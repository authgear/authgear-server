package sso

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestAuthURLPayload(t *testing.T) {
	Convey("Test AuthURLRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := AuthURLRequestPayload{
				CallbackURL:     "callbackURL",
				UXMode:          sso.UXModeWebRedirect,
				OnUserDuplicate: model.OnUserDuplicateAbort,
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without callback url", func() {
			payload := AuthURLRequestPayload{
				UXMode:          sso.UXModeWebRedirect,
				OnUserDuplicate: model.OnUserDuplicateAbort,
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without UX mode", func() {
			payload := AuthURLRequestPayload{
				CallbackURL:     "callbackURL",
				OnUserDuplicate: model.OnUserDuplicateAbort,
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without OnUserDuplicate", func() {
			payload := AuthURLRequestPayload{
				CallbackURL: "callbackURL",
				UXMode:      sso.UXModeWebRedirect,
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
		h.APIClientConfigurationProvider = apiclientconfig.NewMockProvider("api_key")
		h.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
		oauthConfig := coreconfig.OAuthConfiguration{
			StateJWTSecret: "secret",
			AllowedCallbackURLs: []string{
				"callbackURL",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           "mock",
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
			Scope:        "openid profile email",
		}
		mockProvider := sso.MockSSOProvider{
			URLPrefix:      &url.URL{Scheme: "https", Host: "localhost:3000"},
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
		}
		mockPasswordProvider := password.NewMockProvider(
			nil,
			[]string{password.DefaultRealm},
		)
		h.TxContext = db.NewMockTxContext()
		h.Provider = &mockProvider
		h.PasswordAuthProvider = mockPasswordProvider
		h.Action = "login"
		h.OAuthConfiguration = oauthConfig

		Convey("should return login_auth_url", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "callbackURL",
				"ux_mode": "web_redirect"
			}
			`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			httpHandler := h
			httpHandler.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			body := map[string]interface{}{}
			_ = json.NewDecoder(bytes.NewReader(resp.Body.Bytes())).Decode(&body)

			// check base url
			u, _ := url.Parse(body["result"].(string))
			So(u.Host, ShouldEqual, "mock")
			So(u.Path, ShouldEqual, "/auth")

			// check querys
			q := u.Query()
			So(q.Get("response_type"), ShouldEqual, "code")
			So(q.Get("client_id"), ShouldEqual, "mock_client_id")
			So(q.Get("scope"), ShouldEqual, "openid profile email")

			// check redirect_uri
			r, _ := url.Parse(q.Get("redirect_uri"))
			So(r.Host, ShouldEqual, "localhost:3000")
			So(r.Path, ShouldEqual, "/_auth/sso/mock/auth_handler")

			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			_, err := jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
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
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "callbackURL",
				"ux_mode": "web_redirect"
			}
			`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			httpHandler := h
			httpHandler.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			body := map[string]interface{}{}
			_ = json.NewDecoder(bytes.NewReader(resp.Body.Bytes())).Decode(&body)

			// check base url
			u, _ := url.Parse(body["result"].(string))
			q := u.Query()
			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			})
			So(claims.State.Action, ShouldEqual, "link")
		})

		Convey("should reject invalid realm", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "http://localhost:3000",
				"ux_mode": "web_popup",
				"merge_realm": "nonsense"
			}
			`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			httpHandler := h
			httpHandler.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 107,
					"info": {
						"arguments": [
							"nonsense"
						]
					},
					"message": "Invalid MergeRealm",
					"name": "InvalidArgument"
				}
			}
			`)
		})

		Convey("should reject disallowed OnUserDuplicate", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "http://localhost:3000",
				"ux_mode": "web_popup",
				"on_user_duplicate": "merge"
			}
			`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			httpHandler := h
			httpHandler.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 107,
					"info": {
						"arguments": [
							"merge"
						]
					},
					"message": "Disallowed OnUserDuplicate",
					"name": "InvalidArgument"
				}
			}
			`)
		})
	})
}
