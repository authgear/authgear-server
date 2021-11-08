package handler_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	sessiontest "github.com/authgear/authgear-server/pkg/lib/session/test"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const htmlRedirectTemplateString = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url={{ .redirect_uri }}" />
</head>
<body>
<script>
window.location.href = "{{ .redirect_uri }}"
</script>
</body>
</html>
`

func TestAuthorizationHandler(t *testing.T) {

	htmlRedirectTemplate, _ := template.New("html_redirect").Parse(htmlRedirectTemplateString)
	redirectHTML := func(redirectURI string) string {
		buf := strings.Builder{}
		_ = htmlRedirectTemplate.Execute(&buf, map[string]string{
			"redirect_uri": redirectURI,
		})
		return buf.String()
	}
	redirection := func(resp *httptest.ResponseRecorder) string {
		return resp.Header().Get("Location")
	}

	Convey("Authorization handler", t, func() {
		clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")
		authzStore := &mockAuthzStore{}
		codeGrantStore := &mockCodeGrantStore{}
		authenticationInfoService := &mockAuthenticationInfoService{}
		cookieManager := &mockCookieManager{}

		h := &handler.AuthorizationHandler{
			Context: context.Background(),
			AppID:   "app-id",
			Config:  &config.OAuthConfig{},
			HTTPConfig: &config.HTTPConfig{
				PublicOrigin: "http://accounts.example.com",
			},

			Authorizations:            authzStore,
			CodeGrants:                codeGrantStore,
			OAuthURLs:                 mockURLsProvider{},
			WebAppURLs:                mockURLsProvider{},
			ValidateScopes:            func(*config.OAuthClientConfig, []string) error { return nil },
			CodeGenerator:             func() string { return "authz-code" },
			Clock:                     clock,
			AuthenticationInfoService: authenticationInfoService,
			Cookies:                   cookieManager,
		}
		handle := func(r protocol.AuthorizationRequest) *httptest.ResponseRecorder {
			result := h.Handle(r)
			req, _ := http.NewRequest("GET", "/authorize", nil)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}

		Convey("general request validation", func() {
			h.Config.Clients = []config.OAuthClientConfig{{
				ClientID: "client-id",
				RedirectURIs: []string{
					"https://example.com/",
					"https://example.com/settings",
				},
			}}
			Convey("missing client ID", func() {
				resp := handle(protocol.AuthorizationRequest{})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
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
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: invalid_request\n"+
						"error_description: redirect URI is not allowed\n")
			})
			Convey("implicitly allowed redirect URI on AS", func() {
				resp := handle(protocol.AuthorizationRequest{
					"client_id":    "client-id",
					"redirect_uri": "http://accounts.example.com/settings",
				})
				So(resp.Result().StatusCode, ShouldEqual, 200)
				So(resp.Body.String(), ShouldEqual, redirectHTML(
					"http://accounts.example.com/settings?error=unauthorized_client&error_description=response+type+is+not+allowed+for+this+client",
				))
			})
		})

		Convey("should preserve query parameters in redirect URI", func() {
			h.Config.Clients = []config.OAuthClientConfig{{
				ClientID:     "client-id",
				RedirectURIs: []string{"https://example.com/cb?from=sso"},
			}}
			resp := handle(protocol.AuthorizationRequest{
				"client_id":     "client-id",
				"response_type": "code",
			})
			So(resp.Result().StatusCode, ShouldEqual, 200)
			So(resp.Body.String(), ShouldEqual, redirectHTML(
				"https://example.com/cb?error=invalid_request&error_description=scope+is+required&from=sso",
			))
		})

		Convey("authorization code flow", func() {
			h.Config.Clients = []config.OAuthClientConfig{{
				ClientID:     "client-id",
				RedirectURIs: []string{"https://example.com/"},
			}}
			Convey("request validation", func() {
				Convey("missing scope", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=scope+is+required",
					))
				})
				Convey("missing PKCE code challenge", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
						"scope":         "openid",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=PKCE+code+challenge+is+required",
					))
				})
				Convey("unsupported PKCE transform", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid",
						"code_challenge_method": "plain",
						"code_challenge":        "code-verifier",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=only+%27S256%27+PKCE+transform+is+supported",
					))
				})
			})
			Convey("scope validation", func() {
				validated := false
				h.ValidateScopes = func(client *config.OAuthClientConfig, scopes []string) error {
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
				So(resp.Result().StatusCode, ShouldEqual, 200)
				So(resp.Body.String(), ShouldEqual, redirectHTML(
					"https://example.com/?error=invalid_scope&error_description=must+request+%27openid%27+scope",
				))
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
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("return authorization code", func() {
				h.Context = sessiontest.NewMockSession().
					SetUserID("user-id").
					SetSessionID("session-id").
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
						"prompt":                "none",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code&state=my-state",
					))

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
						IDPSessionID:    "session-id",
						AuthenticationInfo: authenticationinfo.T{
							UserID: "user-id",
						},
						CreatedAt:     time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:      time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						Scopes:        []string{"openid"},
						CodeHash:      "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:   "https://example.com/",
						OIDCNonce:     "my-nonce",
						PKCEChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
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
						"prompt":                "none",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code",
					))

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
						IDPSessionID:    "session-id",
						AuthenticationInfo: authenticationinfo.T{
							UserID: "user-id",
						},
						CreatedAt:     time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:      time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						Scopes:        []string{"openid", "offline_access"},
						CodeHash:      "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:   "https://example.com/",
						OIDCNonce:     "",
						PKCEChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					})
				})
			})
		})
		Convey("none response type", func() {
			h.Config.Clients = []config.OAuthClientConfig{{
				ClientID:      "client-id",
				RedirectURIs:  []string{"https://example.com/"},
				ResponseTypes: []string{"none"},
			}}
			Convey("request validation", func() {
				Convey("not allowed response types", func() {
					h.Config.Clients[0].ResponseTypes = nil
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "none",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=unauthorized_client&error_description=response+type+is+not+allowed+for+this+client",
					))
				})
			})
			Convey("scope validation", func() {
				validated := false
				h.ValidateScopes = func(client *config.OAuthClientConfig, scopes []string) error {
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
				So(resp.Result().StatusCode, ShouldEqual, 200)
				So(resp.Body.String(), ShouldEqual, redirectHTML(
					"https://example.com/?error=invalid_scope&error_description=must+request+%27openid%27+scope",
				))
			})
			Convey("request authentication", func() {
				resp := handle(protocol.AuthorizationRequest{
					"client_id":     "client-id",
					"response_type": "none",
					"scope":         "openid",
				})
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("redirect to URI", func() {
				h.Context = sessiontest.NewMockSession().
					SetUserID("user-id").
					SetSessionID("session-id").
					ToContext(context.Background())

				Convey("create new authorization implicitly", func() {
					resp := handle(protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "none",
						"scope":         "openid",
						"state":         "my-state",
						"prompt":        "none",
					})
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?state=my-state",
					))

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
