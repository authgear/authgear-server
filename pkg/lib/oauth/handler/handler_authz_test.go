package handler_test

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")
		authzService := NewMockAuthorizationService(ctrl)
		uiInfoResolver := NewMockUIInfoResolver(ctrl)
		uiURLBuilder := NewMockUIURLBuilder(ctrl)
		codeGrantStore := &mockCodeGrantStore{}
		authenticationInfoService := &mockAuthenticationInfoService{}
		cookieManager := &mockCookieManager{}
		oauthSessionService := &mockOAuthSessionService{}
		clientResolver := &mockClientResolver{}
		appInitiatedSSOToWebTokenService := NewMockAppInitiatedSSOToWebTokenService(ctrl)
		idTokenIssuer := NewMockIDTokenIssuer(ctrl)

		appID := config.AppID("app-id")
		h := &handler.AuthorizationHandler{
			Context:    context.Background(),
			AppID:      appID,
			Config:     &config.OAuthConfig{},
			HTTPOrigin: "http://accounts.example.com",

			UIURLBuilder:              uiURLBuilder,
			UIInfoResolver:            uiInfoResolver,
			Authorizations:            authzService,
			ValidateScopes:            func(*config.OAuthClientConfig, []string) error { return nil },
			Clock:                     clock,
			AuthenticationInfoService: authenticationInfoService,
			Cookies:                   cookieManager,
			OAuthSessionService:       oauthSessionService,
			CodeGrantService: handler.CodeGrantService{
				AppID:         appID,
				Clock:         clock,
				CodeGenerator: func() string { return "authz-code" },
				CodeGrants:    codeGrantStore,
			},
			ClientResolver:                   clientResolver,
			AppInitiatedSSOToWebTokenService: appInitiatedSSOToWebTokenService,
			IDTokenIssuer:                    idTokenIssuer,
		}
		handle := func(r protocol.AuthorizationRequest) *httptest.ResponseRecorder {
			result := h.Handle(r)
			req, _ := http.NewRequest("GET", "/authorize", nil)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}

		Convey("general request validation", func() {
			clientResolver.ClientConfig = &config.OAuthClientConfig{
				ClientID: "client-id",
				RedirectURIs: []string{
					"https://example.com/",
					"https://example.com/settings",
				},
				CustomUIURI: "https://ui.custom.com",
			}
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
			clientResolver.ClientConfig = &config.OAuthClientConfig{
				ClientID:     "client-id",
				RedirectURIs: []string{"https://example.com/cb?from=sso"},
				CustomUIURI:  "https://ui.custom.com",
			}
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
			mockedClient := &config.OAuthClientConfig{
				ClientID:     "client-id",
				RedirectURIs: []string{"https://example.com/"},
				CustomUIURI:  "https://ui.custom.com",
			}
			clientResolver.ClientConfig = mockedClient
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
						"https://example.com/?error=invalid_request&error_description=PKCE+code+challenge+is+required+for+public+clients",
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
				req := protocol.AuthorizationRequest{
					"client_id":             "client-id",
					"response_type":         "code",
					"scope":                 "openid",
					"code_challenge_method": "S256",
					"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					"ui_locales":            "ja",
				}
				uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
					mockedClient,
					req,
				).Times(1).Return(&oidc.UIInfo{}, &oidc.UIInfoByProduct{}, nil)
				uiURLBuilder.EXPECT().BuildAuthenticationURL(mockedClient, req, gomock.Any()).Times(1).Return(&url.URL{
					Scheme: "https",
					Host:   "auth",
					Path:   "/authenticate",
				}, nil)
				resp := handle(req)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("return authorization code", func() {
				h.Context = sessiontest.NewMockSession().
					SetUserID("user-id").
					SetSessionID("session-id").
					ToContext(context.Background())

				Convey("create new authorization implicitly", func() {
					req := protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid",
						"code_challenge_method": "S256",
						"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
						"nonce":                 "my-nonce",
						"state":                 "my-state",
						"prompt":                "none",
					}
					authorization := &oauth.Authorization{
						ID:        "authz-id",
						AppID:     string(appID),
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid"},
					}
					uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)
					authzService.EXPECT().CheckAndGrant(
						"client-id",
						"user-id",
						[]string{"openid"},
					).Times(1).Return(authorization, nil)

					resp := handle(req)
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code&state=my-state",
					))

					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: authorization.ID,
						IDPSessionID:    "session-id",
						AuthenticationInfo: authenticationinfo.T{
							UserID: "user-id",
						},
						CreatedAt:            time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:             time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						CodeHash:             "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:          "https://example.com/",
						AuthorizationRequest: req,
					})
				})

				Convey("reuse existing authorization implicitly", func() {
					authorization := &oauth.Authorization{
						ID:        "authz-id",
						AppID:     string(appID),
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid", "offline_access"},
					}
					authzService.EXPECT().CheckAndGrant(
						"client-id",
						"user-id",
						[]string{"openid", "offline_access"},
					).Times(1).Return(authorization, nil)
					req := protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid offline_access",
						"code_challenge_method": "S256",
						"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
						"prompt":                "none",
					}
					uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)

					resp := handle(req)
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code",
					))

					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: "authz-id",
						IDPSessionID:    "session-id",
						AuthenticationInfo: authenticationinfo.T{
							UserID: "user-id",
						},
						CreatedAt:            time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						ExpireAt:             time.Date(2020, 2, 1, 0, 5, 0, 0, time.UTC),
						CodeHash:             "f70a35079d7afc23fc5cff56bcd1430b7ce75cd19eaa41132076715b1cea104a",
						RedirectURI:          "https://example.com/",
						AuthorizationRequest: req,
					})
				})
			})
		})
		Convey("none response type", func() {
			mockedClient := &config.OAuthClientConfig{
				ClientID:      "client-id",
				RedirectURIs:  []string{"https://example.com/"},
				ResponseTypes: []string{"none"},
				CustomUIURI:   "https://ui.custom.com",
			}
			clientResolver.ClientConfig = mockedClient
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
				req := protocol.AuthorizationRequest{
					"client_id":     "client-id",
					"response_type": "none",
					"scope":         "openid",
				}
				uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
					mockedClient,
					req,
				).Times(1).Return(&oidc.UIInfo{}, &oidc.UIInfoByProduct{}, nil)
				uiURLBuilder.EXPECT().BuildAuthenticationURL(mockedClient, req, gomock.Any()).Times(1).Return(&url.URL{
					Scheme: "https",
					Host:   "auth",
					Path:   "/authenticate",
				}, nil)
				resp := handle(req)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("redirect to URI", func() {
				h.Context = sessiontest.NewMockSession().
					SetUserID("user-id").
					SetSessionID("session-id").
					ToContext(context.Background())

				Convey("create new authorization implicitly", func() {
					authorization := &oauth.Authorization{
						ID:        "authz-id",
						AppID:     string(appID),
						ClientID:  "client-id",
						UserID:    "user-id",
						CreatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
						Scopes:    []string{"openid"},
					}
					req := protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "none",
						"scope":         "openid",
						"state":         "my-state",
						"prompt":        "none",
					}
					authzService.EXPECT().CheckAndGrant(
						"client-id",
						"user-id",
						[]string{"openid"},
					).Times(1).Return(authorization, nil)
					uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)

					resp := handle(req)
					So(resp.Result().StatusCode, ShouldEqual, 200)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?state=my-state",
					))

					So(codeGrantStore.grants, ShouldBeEmpty)
				})
			})
		})

		Convey("app-initiated-sso-to-web", func() {
			mockedClient := &config.OAuthClientConfig{
				ClientID:      "client-id",
				RedirectURIs:  []string{"https://example.com/"},
				ResponseTypes: []string{"none", "urn:authgear:params:oauth:response-type:app_initiated_sso_to_web token"},
			}
			clientResolver.ClientConfig = mockedClient

			Convey("exchange for access token in cookie", func() {
				testOfflineGrantID := "TEST_OFFLINE_GRANT_ID"
				testOfflineGrant := &oauth.OfflineGrant{
					ID: testOfflineGrantID,
				}
				testSID := oidc.EncodeSID(testOfflineGrant)

				testAppInititatedSSOToWebToken := "TEST_APP_INITIATED_SSO_TO_WEB_TOKEN"
				testIDToken := "TEST_ID_TOKEN"

				testVerifiedIDToken := jwt.New()
				testVerifiedIDToken.Set(string(model.ClaimSID), testSID)

				idTokenIssuer.EXPECT().VerifyIDTokenWithoutClient(testIDToken).
					Times(1).
					Return(testVerifiedIDToken, nil)

				testAccessToken := "TEST_ACCESS_TOKEN"

				appInitiatedSSOToWebTokenService.EXPECT().ExchangeForAccessToken(
					mockedClient,
					testOfflineGrantID,
					testAppInititatedSSOToWebToken,
				).
					Times(1).
					Return(testAccessToken, nil)

				req := protocol.AuthorizationRequest{
					"client_id":                        "client-id",
					"response_type":                    "urn:authgear:params:oauth:response-type:app_initiated_sso_to_web token",
					"x_app_initiated_sso_to_web_token": testAppInititatedSSOToWebToken,
					"prompt":                           "none",
					"response_mode":                    "cookie",
					"state":                            "my-state",
					"redirect_uri":                     "https://example.com/",
					"id_token_hint":                    testIDToken,
				}

				resp := handle(req)
				So(resp.Result().StatusCode, ShouldEqual, 200)
				So(resp.Body.String(), ShouldEqual, redirectHTML(
					"https://example.com/?state=my-state",
				))
				cookieSet := false
				for _, cookie := range resp.Result().Cookies() {
					if cookie.Name == "app_access_token" && cookie.Value == testAccessToken {
						cookieSet = true
					}
				}
				So(cookieSet, ShouldEqual, true)

			})
		})
	})
}
