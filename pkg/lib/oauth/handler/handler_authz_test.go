package handler_test

import (
	"context"
	"errors"
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
	"github.com/authgear/authgear-server/pkg/lib/cimd"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	sessiontest "github.com/authgear/authgear-server/pkg/lib/session/test"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// orderCheckingResolver lets a test observe that CIMDService.EnsureClientResolved
// ran to completion before ClientResolver.ResolveClient is invoked.
type orderCheckingResolver struct {
	inner     handler.OAuthClientResolver
	onResolve func()
}

func (r *orderCheckingResolver) ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig {
	r.onResolve()
	return r.inner.ResolveClient(ctx, clientID)
}

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
		clientResolver := &multiClientResolver{
			ClientConfigs: make(map[string]*config.OAuthClientConfig),
		}
		preAuthenticatedURLTokenService := NewMockAuthorizationHandlerPreAuthenticatedURLTokenService(ctrl)
		idTokenIssuer := NewMockIDTokenIssuer(ctrl)
		accessTokenEncoding := NewMockAuthorizationHandlerAccessTokenEncoding(ctrl)
		// cimdEnsureFn defaults to a no-op (nil): every existing test in this
		// file uses a static-shaped client_id, so CIMD resolution is always
		// a no-op -- this is the regression guard that CIMD wiring doesn't
		// disturb any existing flow. A single AnyTimes() expectation
		// delegates to this reassignable func, rather than each leaf test
		// registering its own EXPECT() -- gomock matches expected calls in
		// the order they were recorded, so a second, more specific
		// expectation added inside a nested Convey would never be reached
		// behind an earlier catch-all AnyTimes() one.
		cimdEnsureFn := func(ctx context.Context, clientID string) error { return nil }
		cimdService := NewMockAuthorizationHandlerCIMDService(ctrl)
		cimdService.EXPECT().EnsureClientResolved(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
			func(ctx context.Context, clientID string) error {
				return cimdEnsureFn(ctx, clientID)
			})

		appID := config.AppID("app-id")
		h := &handler.AuthorizationHandler{
			AppID:      appID,
			Config:     &config.OAuthConfig{},
			HTTPOrigin: "http://accounts.example.com",

			Database: &db.MockHandle{},

			UIURLBuilder:              uiURLBuilder,
			UIInfoResolver:            uiInfoResolver,
			Authorizations:            authzService,
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
			ClientResolver:                          clientResolver,
			PreAuthenticatedURLTokenService:         preAuthenticatedURLTokenService,
			IDTokenIssuer:                           idTokenIssuer,
			AuthorizationHandlerAccessTokenEncoding: accessTokenEncoding,
			CIMDService:                             cimdService,
		}
		handle := func(ctx context.Context, r protocol.AuthorizationRequest) *httptest.ResponseRecorder {
			ctx, params, errResult := h.ValidateRequestWithoutTx(ctx, r)
			var result httputil.Result
			if errResult != nil {
				result = errResult
			} else {
				result = h.HandleRequest(ctx, r, params)
			}

			req, _ := http.NewRequest("GET", "/authorize", nil)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}

		Convey("general request validation", func() {
			clientResolver.ClientConfigs["client-id"] = &config.OAuthClientConfig{
				ClientID: "client-id",
				RedirectURIs: []string{
					"https://example.com/",
					"https://example.com/settings",
				},
				CustomUIURI: "https://ui.custom.com",
			}
			Convey("missing client ID", func() {
				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: unauthorized_client\n"+
						"error_description: invalid client ID\n")
			})
			Convey("disallowed redirect URI", func() {
				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
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
				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id":    "client-id",
					"redirect_uri": "http://accounts.example.com/settings",
				})
				So(resp.Body.String(), ShouldEqual, redirectHTML(
					"http://accounts.example.com/settings?error=unauthorized_client&error_description=response+type+is+not+allowed+for+this+client",
				))
			})
		})

		Convey("CIMD resolution", func() {
			Convey("EnsureClientResolved is called before ClientResolver.ResolveClient", func() {
				ensureCalled := false
				cimdEnsureFn = func(ctx context.Context, clientID string) error {
					So(clientID, ShouldEqual, "client-id")
					ensureCalled = true
					return nil
				}
				clientResolver.ClientConfigs["client-id"] = &config.OAuthClientConfig{
					ClientID:     "client-id",
					RedirectURIs: []string{"https://example.com/"},
				}
				h.ClientResolver = &orderCheckingResolver{
					inner: clientResolver,
					onResolve: func() {
						So(ensureCalled, ShouldBeTrue)
					},
				}

				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id":    "client-id",
					"redirect_uri": "https://example.com/",
				})
				So(resp.Result().StatusCode, ShouldNotEqual, 500)
				So(ensureCalled, ShouldBeTrue)
			})

			Convey("CIMDUnresolvable produces the byte-identical unknown-client_id response", func() {
				cimdEnsureFn = func(ctx context.Context, clientID string) error {
					return cimd.ErrUnresolvable()
				}

				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
				})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: unauthorized_client\n"+
						"error_description: invalid client ID\n")

				// Byte-identical to the ordinary unknown-client_id response
				// (spec § Authgear as an SSRF/Probing Oracle): compare
				// against the existing "missing client ID" case's body.
				missing := handle(context.Background(), protocol.AuthorizationRequest{})
				So(resp.Body.String(), ShouldEqual, missing.Body.String())
			})

			Convey("CIMDClientLimitExceeded maps to access_denied", func() {
				cimdEnsureFn = func(ctx context.Context, clientID string) error {
					return cimd.ErrClientLimitExceeded()
				}

				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
				})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: access_denied\n"+
						"error_description: the project has reached its client limit\n")
			})

			Convey("a ratelimit.RateLimited error maps to x_rate_limited", func() {
				cimdEnsureFn = func(ctx context.Context, clientID string) error {
					return ratelimit.ErrRateLimited("", "", "")
				}

				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
				})
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: x_rate_limited\n"+
						"error_description: rate limit exceeded, please try again later.\n")
			})

			Convey("any other error maps to server_error", func() {
				cimdEnsureFn = func(ctx context.Context, clientID string) error {
					return errors.New("boom")
				}

				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
				})
				// AuthorizationResultError with no RedirectURI always
				// renders as a 400 with the error body -- InternalError
				// only marks it for logging/sentry (IsInternalError()), it
				// is not the HTTP status.
				So(resp.Result().StatusCode, ShouldEqual, 400)
				So(resp.Body.String(), ShouldEqual,
					"Invalid OAuth authorization request:\n"+
						"error: server_error\n"+
						"error_description: internal server error\n")
			})
		})

		Convey("should preserve query parameters in redirect URI", func() {
			clientResolver.ClientConfigs["client-id"] = &config.OAuthClientConfig{
				ClientID:     "client-id",
				RedirectURIs: []string{"https://example.com/cb?from=sso"},
				CustomUIURI:  "https://ui.custom.com",
			}
			ctx := context.Background()
			resp := handle(ctx, protocol.AuthorizationRequest{
				"client_id":     "client-id",
				"response_type": "code",
			})
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
			clientResolver.ClientConfigs["client-id"] = mockedClient
			Convey("request validation", func() {
				Convey("missing scope", func() {
					ctx := context.Background()
					resp := handle(ctx, protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
					})
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=scope+is+required",
					))
				})
				Convey("missing PKCE code challenge", func() {
					ctx := context.Background()
					resp := handle(ctx, protocol.AuthorizationRequest{
						"client_id":     "client-id",
						"response_type": "code",
						"scope":         "openid",
					})
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=PKCE+code+challenge+is+required+for+public+clients",
					))
				})
				Convey("unsupported PKCE transform", func() {
					ctx := context.Background()
					resp := handle(ctx, protocol.AuthorizationRequest{
						"client_id":             "client-id",
						"response_type":         "code",
						"scope":                 "openid",
						"code_challenge_method": "plain",
						"code_challenge":        "code-verifier",
					})
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?error=invalid_request&error_description=only+%27S256%27+PKCE+transform+is+supported",
					))
				})
			})
			Convey("scope validation", func() {
				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id":             "client-id",
					"response_type":         "code",
					"scope":                 "email",
					"code_challenge_method": "S256",
					"code_challenge":        "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				})
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
					gomock.Any(),
					mockedClient,
					req,
				).Times(1).Return(&oidc.UIInfo{}, &oidc.UIInfoByProduct{}, nil)
				uiURLBuilder.EXPECT().BuildAuthenticationURL(mockedClient, req, gomock.Any()).Times(1).Return(&url.URL{
					Scheme: "https",
					Host:   "auth",
					Path:   "/authenticate",
				}, nil)
				ctx := context.Background()
				resp := handle(ctx, req)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("return authorization code", func() {
				ctx := sessiontest.NewMockSession().
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
						gomock.Any(),
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)
					authzService.EXPECT().CheckAndGrant(
						gomock.Any(),
						"client-id",
						"user-id",
						[]string{"openid"},
					).Times(1).Return(authorization, nil)

					resp := handle(ctx, req)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code&state=my-state",
					))

					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: authorization.ID,
						AuthenticationInfo: authenticationinfo.T{
							UserID:                     "user-id",
							AuthenticatedBySessionID:   "session-id",
							AuthenticatedBySessionType: string(session.TypeIdentityProvider),
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
						gomock.Any(),
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
						gomock.Any(),
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)

					resp := handle(ctx, req)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?code=authz-code",
					))

					So(codeGrantStore.grants, ShouldHaveLength, 1)
					So(codeGrantStore.grants[0], ShouldResemble, oauth.CodeGrant{
						AppID:           "app-id",
						AuthorizationID: "authz-id",
						AuthenticationInfo: authenticationinfo.T{
							UserID:                     "user-id",
							AuthenticatedBySessionID:   "session-id",
							AuthenticatedBySessionType: string(session.TypeIdentityProvider),
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
			clientResolver.ClientConfigs["client-id"] = mockedClient
			Convey("scope validation", func() {
				ctx := context.Background()
				resp := handle(ctx, protocol.AuthorizationRequest{
					"client_id":     "client-id",
					"response_type": "none",
					"scope":         "email",
				})
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
					gomock.Any(),
					mockedClient,
					req,
				).Times(1).Return(&oidc.UIInfo{}, &oidc.UIInfoByProduct{}, nil)
				uiURLBuilder.EXPECT().BuildAuthenticationURL(mockedClient, req, gomock.Any()).Times(1).Return(&url.URL{
					Scheme: "https",
					Host:   "auth",
					Path:   "/authenticate",
				}, nil)
				ctx := context.Background()
				resp := handle(ctx, req)
				So(resp.Result().StatusCode, ShouldEqual, 302)
				So(redirection(resp), ShouldEqual, "https://auth/authenticate")
			})
			Convey("redirect to URI", func() {
				ctx := sessiontest.NewMockSession().
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
						gomock.Any(),
						"client-id",
						"user-id",
						[]string{"openid"},
					).Times(1).Return(authorization, nil)
					uiInfoResolver.EXPECT().ResolveForAuthorizationEndpoint(
						gomock.Any(),
						mockedClient,
						req,
					).Times(1).Return(&oidc.UIInfo{
						Prompt: []string{"none"},
					}, &oidc.UIInfoByProduct{}, nil)

					resp := handle(ctx, req)
					So(resp.Body.String(), ShouldEqual, redirectHTML(
						"https://example.com/?state=my-state",
					))

					So(codeGrantStore.grants, ShouldBeEmpty)
				})
			})
		})

		Convey("pre-authenticated-url", func() {
			mockedClient := &config.OAuthClientConfig{
				ClientID:      "client-id",
				RedirectURIs:  []string{"https://example.com/"},
				ResponseTypes: []string{"none", "urn:authgear:params:oauth:response-type:pre-authenticated-url token"},
			}
			clientResolver.ClientConfigs["client-id"] = mockedClient

			Convey("exchange for access token in cookie", func() {
				testOfflineGrantID := "TEST_OFFLINE_GRANT_ID"
				testOfflineGrant := &oauth.OfflineGrant{
					ID: testOfflineGrantID,
				}
				testSID := oauth.EncodeSID(testOfflineGrant)

				// nolint:gosec
				testPreAuthenticatedURLToken := "TEST_PRE_AUTHENTICATED_URL_TOKEN"
				testIDToken := "TEST_ID_TOKEN"

				testVerifiedIDToken := jwt.New()
				_ = testVerifiedIDToken.Set(string(model.ClaimSID), testSID)

				idTokenIssuer.EXPECT().VerifyIDToken(testIDToken).
					Times(1).
					Return(testVerifiedIDToken, nil)

				testAccessToken := "TEST_ACCESS_TOKEN"

				preAuthenticatedURLTokenService.EXPECT().ExchangeForAccessToken(
					gomock.Any(),
					mockedClient,
					testOfflineGrantID,
					testPreAuthenticatedURLToken,
				).
					Times(1).
					Return(nil, nil)
				accessTokenEncoding.EXPECT().MakeUserAccessTokenFromPreparationResult(gomock.Any(), gomock.Any()).Times(1).Return(&oauth.IssueAccessGrantResult{
					Token: testAccessToken,
				}, nil)

				req := protocol.AuthorizationRequest{
					"client_id":                     "client-id",
					"response_type":                 "urn:authgear:params:oauth:response-type:pre-authenticated-url token",
					"x_pre_authenticated_url_token": testPreAuthenticatedURLToken,
					"prompt":                        "none",
					"response_mode":                 "cookie",
					"state":                         "my-state",
					"redirect_uri":                  "https://example.com/",
					"id_token_hint":                 testIDToken,
				}

				ctx := context.Background()
				resp := handle(ctx, req)
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
