package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthorizationHandler(t *testing.T) {
	Convey("Authorization handler", t, func() {
		mockTime := &coretime.MockProvider{}
		mockTime.TimeNowUTC = time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)
		authzStore := &mockAuthzStore{}
		codeGrantStore := &mockCodeGrantStore{}

		h := &handler.AuthorizationHandler{
			Context: context.Background(),
			AppID:   "app-id",
			Clients: []config.OAuthClientConfiguration{},

			Authorizations:  authzStore,
			CodeGrants:      codeGrantStore,
			AuthorizeURL:    mockEndpointsProvider{},
			AuthenticateURL: mockEndpointsProvider{},
			ValidateScopes:  func(config.OAuthClientConfiguration, []string) error { return nil },
			CodeGenerator:   func() string { return "authz-code" },
			Time:            mockTime,
		}
		handle := func(r protocol.AuthorizationRequest) *httptest.ResponseRecorder {
			result := h.Handle(r)
			req, _ := http.NewRequest("GET", "/authorize", nil)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}
		redirection := func(resp *httptest.ResponseRecorder) string {
			return resp.Header().Get("Location")
		}

		Convey("general request validation", func() {
			h.Clients = []config.OAuthClientConfiguration{{
				"client_id": "client-id",
				"redirect_uris": []interface{}{
					"https://example.com/",
					"https://example.com/settings",
				},
			}}
			Convey("missing client ID", func() {
				resp := handle(protocol.AuthorizationRequest{})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(string(resp.Body.Bytes()), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: unauthorized_client\n"+
						"error_description: invalid client ID\n")
			})
			Convey("disallowed redirect URI", func() {
				resp := handle(protocol.AuthorizationRequest{
					"client_id":    "client-id",
					"redirect_uri": "https://example.com",
				})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(string(resp.Body.Bytes()), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: invalid_request\n"+
						"error_description: redirect URI is not allowed\n")
			})
		})

		Convey("should preserve query parameters in redirect URI", func() {
			h.Clients = []config.OAuthClientConfiguration{{
				"client_id":     "client-id",
				"redirect_uris": []interface{}{"https://example.com/cb?from=sso"},
			}}
			resp := handle(protocol.AuthorizationRequest{
				"client_id":     "client-id",
				"response_type": "code",
			})
			So(resp.Result().StatusCode, ShouldEqual, 302)
			So(redirection(resp), ShouldEqual,
				"https://example.com/cb?error=invalid_request&error_description=scope+is+required&from=sso")
		})

		Convey("authorization code flow", func() {
			h.Clients = []config.OAuthClientConfiguration{{
				"client_id":     "client-id",
				"redirect_uris": []interface{}{"https://example.com/"},
			}}
			Convey("request validation", func() {
				Convey("missing scope", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?error=invalid_request&error_description=scope+is+required")
				})
				Convey("missing PKCE code challenge", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
						"scope":         "openid",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?error=invalid_request&error_description=PKCE+code+challenge+is+required")
				})
				Convey("unsupported PKCE transform", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid",
						"code_challenge_method": "plain",
						"code_challenge":        "code-verifier",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?error=invalid_request&error_description=only+%27S256%27+PKCE+transform+is+supported")
				})
			})
			Convey("scope validation", func() {
				validated := false
				h.ValidateScopes = func(client config.OAuthClientConfiguration, scopes []string) error {
					validated = true
					if strings.Join(scopes, " ") != "openid" {
						return protocol.NewError("invalid_scope", "must request 'openid' scope")
					}
					return nil
				}

				resp := handle(protocol.AuthorizationRequest{
					"client_id":             "client-id",
					"response_type":         "code",
					"scope":                 "email",
					"code_challenge_method": "S256",
					"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				})
				So(validated, ShouldBeTrue)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual,
					"https://example.com/?error=invalid_scope&error_description=must+request+%27openid%27+scope")
			})
			Convey("request authentication", func() {
				resp := handle(protocol.AuthorizationRequest{
					"client_id":             "client-id",
					"response_type":         "code",
					"scope":                 "openid",
					"code_challenge_method": "S256",
					"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					"ui_locales":            "ja",
				})
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual,
					"https://auth/authenticate?client_id=client-id&redirect_uri=https%3A%2F%2Fauth%2Fauthorize%3Fclient_id%3Dclient-id%26code_challenge%3DE9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM%26code_challenge_method%3DS256%26response_type%3Dcode%26scope%3Dopenid%26ui_locales%3Dja&ui_locales=ja")
			})
			Convey("return authorization code", func() {
				h.Context = authtesting.WithAuthn().
					UserID("user-id").
					SessionID("session-id").
					ToContext(context.Background())

				Convey("create new authorization implicitly", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid",
						"code_challenge_method": "S256",
						"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
						"nonce":                 "my-nonce",
						"state":                 "my-state",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?code=authz-code&state=my-state")

					So(authzStore.authzs, ShouldHaveLength, 1)
					So(authzStore.authzs[0], ShouldResemble, oauth.Authorization{
						ID:        authzStore.authzs[0].ID,
						AppID:     "app-id",
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid"},
					})
					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: authzStore.authzs[0].ID,
						SessionID:       "session-id",
						CreatedAt:       time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:        time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						Scopes:          []string{"openid"},
						CodeHash:        "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:     "https://example.com/",
						OIDCNonce:       "my-nonce",
						PKCEChallenge:   "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					})
				})

				Convey("reuse existing authorization implicitly", func() {
					authzStore.authzs = []oauth.Authorization{{
						ID:        "authz-id",
						AppID:     "app-id",
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid"},
					}}

					resp := handle(protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid offline_access",
						"code_challenge_method": "S256",
						"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?code=authz-code")

					So(authzStore.authzs, ShouldHaveLength, 1)
					So(authzStore.authzs[0], ShouldResemble, oauth.Authorization{
						ID:        "authz-id",
						AppID:     "app-id",
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid", "offline_access"},
					})
					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: "authz-id",
						SessionID:       "session-id",
						CreatedAt:       time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:        time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						Scopes:          []string{"openid", "offline_access"},
						CodeHash:        "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:     "https://example.com/",
						OIDCNonce:       "",
						PKCEChallenge:   "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					})
				})
			})
		})
		Convey("none response type", func() {
			h.Clients = []config.OAuthClientConfiguration{{
				"client_id":      "client-id",
				"redirect_uris":  []interface{}{"https://example.com/"},
				"response_types": []interface{}{"none"},
			}}
			Convey("request validation", func() {
				Convey("not allowed response types", func() {
					h.Clients[0]["response_types"] = nil
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "none",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?error=unauthorized_client&error_description=response+type+is+not+allowed+for+this+client")
				})
			})
			Convey("scope validation", func() {
				validated := false
				h.ValidateScopes = func(client config.OAuthClientConfiguration, scopes []string) error {
					validated = true
					if strings.Join(scopes, " ") != "openid" {
						return protocol.NewError("invalid_scope", "must request 'openid' scope")
					}
					return nil
				}

				resp := handle(protocol.AuthorizationRequest{
					"client_id":     "client-id",
					"response_type": "none",
					"scope":         "email",
				})
				So(validated, ShouldBeTrue)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual,
					"https://example.com/?error=invalid_scope&error_description=must+request+%27openid%27+scope")
			})
			Convey("request authentication", func() {
				resp := handle(protocol.AuthorizationRequest{
					"client_id":     "client-id",
					"response_type": "none",
					"scope":         "openid",
				})
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual,
					"https://auth/authenticate?client_id=client-id&redirect_uri=https%3A%2F%2Fauth%2Fauthorize%3Fclient_id%3Dclient-id%26response_type%3Dnone%26scope%3Dopenid")
			})
			Convey("redirect to URI", func() {
				h.Context = authtesting.WithAuthn().
					UserID("user-id").
					SessionID("session-id").
					ToContext(context.Background())

				Convey("create new authorization implicitly", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "none",
						"scope":         "openid",
						"state":         "my-state",
					})
					So(resp.Result().StatusCode, ShouldEqual, 302)
					So(redirection(resp), ShouldEqual,
						"https://example.com/?state=my-state")

					So(authzStore.authzs, ShouldHaveLength, 1)
					So(authzStore.authzs[0], ShouldResemble, oauth.Authorization{
						ID:        authzStore.authzs[0].ID,
						AppID:     "app-id",
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid"},
					})
					So(codeGrantStore.grants, ShouldBeEmpty)
				})
			})
		})
	})
}
