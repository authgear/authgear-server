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
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestAuthURLHandler(t *testing.T) {
	Convey("Test TestAuthURLHandler", t, func() {
		h := &AuthURLHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			AuthURLRequestSchema,
		)
		h.Validator = validator
		h.APIClientConfigurationProvider = apiclientconfig.NewMockProvider("api_key")
		h.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
		oauthConfig := &coreconfig.OAuthConfiguration{
			StateJWTSecret: "secret",
			AllowedCallbackURLs: []string{
				"http://example.com/sso",
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
		h.ProviderID = "google"
		h.PasswordAuthProvider = mockPasswordProvider
		h.Action = "login"
		h.OAuthConfiguration = oauthConfig

		Convey("should return login_auth_url", func() {
			req, _ := http.NewRequest("POST", "http://mock", strings.NewReader(`
			{
				"callback_url": "http://example.com/sso",
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
			So(u.Path, ShouldEqual, "/_auth/sso/google/auth_redirect")

			// check querys
			q := u.Query()

			// check encoded state
			s := q.Get("state")
			claims := sso.CustomClaims{}
			_, err := jwt.ParseWithClaims(s, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			})
			So(err, ShouldBeNil)
			So(claims.State.UXMode, ShouldEqual, sso.UXModeWebRedirect)
			So(claims.State.CallbackURL, ShouldEqual, "http://example.com/sso")
			So(claims.State.Action, ShouldEqual, "login")
			So(claims.State.UserID, ShouldEqual, "faseng.cat.id")
		})

		Convey("should return link_auth_url", func() {
			h.Action = "link"
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "http://example.com/sso",
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

		SkipConvey("should reject invalid realm", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "http://example.com/sso",
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
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "General", "message": "merge_realm is not a valid realm", "pointer": "/merge_realm" }
						]
					}
				}
			}`)
		})

		Convey("should reject disallowed OnUserDuplicate", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"callback_url": "http://example.com/sso",
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
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "General", "message": "on_user_duplicate is not allowed", "pointer": "/on_user_duplicate" }
						]
					}
				}
			}`)
		})
	})
}
