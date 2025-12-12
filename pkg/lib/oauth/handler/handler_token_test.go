package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestTokenHandler(t *testing.T) {

	Convey("Token handler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clientResolver := &multiClientResolver{
			ClientConfigs: make(map[string]*config.OAuthClientConfig),
		}
		origin := "http://accounts.example.com"
		idTokenIssuer := NewMockIDTokenIssuer(ctrl)
		idTokenIssuer.EXPECT().Iss().Return(origin).AnyTimes()

		appDB := NewMockTokenHandlerAppDatabase(ctrl)
		appDB.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		).AnyTimes()

		offlineGrants := NewMockTokenHandlerOfflineGrantStore(ctrl)

		authorizations := NewMockAuthorizationService(ctrl)
		codeGrantService := NewMockTokenHandlerCodeGrantService(ctrl)
		rateLimiter := NewMockTokenHandlerRateLimiter(ctrl)
		challenges := NewMockChallengeProvider(ctrl)
		app2appService := NewMockApp2AppService(ctrl)

		preAuthenticatedURLService := NewMockPreAuthenticatedURLTokenService(ctrl)

		appID := "testapp"

		tokenService := NewMockTokenHandlerTokenService(ctrl)
		accessTokenEncoding := NewMockTokenHandlerAccessTokenEncoding(ctrl)

		offlineGrantService := NewMockTokenHandlerOfflineGrantService(ctrl)
		clientResourceScopeService := NewMockTokenHandlerClientResourceScopeService(ctrl)
		appSessionTokens := NewMockTokenHandlerAppSessionTokenStore(ctrl)

		clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")

		h := handler.TokenHandler{
			Database:               appDB,
			AppID:                  config.AppID(appID),
			AppDomains:             []string{},
			HTTPProto:              "http",
			HTTPOrigin:             httputil.HTTPOrigin(origin),
			OAuthFeatureConfig:     &config.OAuthFeatureConfig{},
			IdentityFeatureConfig:  &config.IdentityFeatureConfig{},
			OAuthClientCredentials: &config.OAuthClientCredentials{},
			Clock:                  clock,
			RemoteIP:               "1.2.3.4",
			UserAgentString:        "UA",

			TokenService:                    tokenService,
			AccessTokenEncoding:             accessTokenEncoding,
			ClientResolver:                  clientResolver,
			Authorizations:                  authorizations,
			OfflineGrants:                   offlineGrants,
			OfflineGrantService:             offlineGrantService,
			IDTokenIssuer:                   idTokenIssuer,
			PreAuthenticatedURLTokenService: preAuthenticatedURLService,
			ClientResourceScopeService:      clientResourceScopeService,
			RateLimiter:                     rateLimiter,
			AppSessionTokens:                appSessionTokens,
			CodeGrantService:                codeGrantService,
			Challenges:                      challenges,
			App2App:                         app2appService,
		}

		handle := func(ctx context.Context, req *http.Request, r protocol.TokenRequest) *httptest.ResponseRecorder {
			result := h.Handle(ctx, &httptest.ResponseRecorder{}, req, r)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}

		Convey("handle refresh token", func() {
			Convey("success", func() {
				req, _ := http.NewRequest("POST", "/token", nil)
				clientResolver.ClientConfigs["app-id"] = &config.OAuthClientConfig{
					ClientID: "app-id",
					RedirectURIs: []string{
						"https://example.com/",
					},
				}
				r := protocol.TokenRequest{}
				r["grant_type"] = []string{"refresh_token"}
				r["client_id"] = []string{"app-id"}
				r["refresh_token"] = []string{"asdf"}
				refreshTokenHash := "hash1"
				offlineGrant := &oauth.OfflineGrant{
					ID:              "offline-grant-id",
					Attrs:           *session.NewAttrs("user-id"),
					InitialClientID: "app-id",
					RefreshTokens: []oauth.OfflineGrantRefreshToken{{
						ClientID:         "app-id",
						Scopes:           []string{"openid"},
						InitialTokenHash: refreshTokenHash,
					}},
					ExpireAtForResolvedSession: time.Date(2020, 02, 01, 1, 0, 0, 0, time.UTC),
				}
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerUser,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerUser,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"user-id"},
					Period:         time.Minute,
					Burst:          60,
					Enabled:        true,
				}).Return(nil, nil)
				tokenService.EXPECT().ParseRefreshToken(gomock.Any(), "asdf").Return(&oauth.Authorization{}, offlineGrant, refreshTokenHash, nil)

				idTokenIssuer.EXPECT().PrepareIDToken(gomock.Any(), gomock.Any()).Return(&oidc.PrepareIDTokenResult{}, nil)
				idTokenIssuer.EXPECT().MakeIDTokenFromPreparationResult(gomock.Any(), gomock.Any()).Return("id-token", nil)

				tokenService.EXPECT().PrepareUserAccessGrantByRefreshToken(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, options handler.PrepareUserAccessGrantByRefreshTokenOptions) (*handler.PrepareUserAccessGrantByRefreshTokenResult, error) {
						result := &handler.PrepareUserAccessGrantByRefreshTokenResult{
							// The value is unimportant.
							PreparationResult: nil,
						}
						if options.ShouldRotateRefreshToken {
							result.RotateRefreshTokenResult = &oauth.RotateRefreshTokenResult{
								Token: "new-refresh-token",
							}
						}

						return result, nil
					})
				accessTokenEncoding.EXPECT().MakeUserAccessTokenFromPreparationResult(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, options oauth.MakeUserAccessTokenFromPreparationOptions) (*oauth.IssueAccessGrantResult, error) {
					return &oauth.IssueAccessGrantResult{
						Token:     "access-token",
						TokenType: "Bearer",
						ExpiresIn: 300,
					}, nil
				})

				event := access.NewEvent(clock.NowUTC(), "1.2.3.4", "UA")
				offlineGrantService.EXPECT().AccessOfflineGrant(gomock.Any(), "offline-grant-id", refreshTokenHash, &event, offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				offlineGrants.EXPECT().UpdateOfflineGrantDeviceInfo(gomock.Any(), "offline-grant-id", gomock.Any(), offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				ctx := context.Background()
				res := handle(ctx, req, r)
				So(res.Result().StatusCode, ShouldEqual, 200)
			})

			Convey("should rotate refresh token if enabled", func() {
				req, _ := http.NewRequest("POST", "/token", nil)
				clientResolver.ClientConfigs["app-id"] = &config.OAuthClientConfig{
					ClientID: "app-id",
					RedirectURIs: []string{
						"https://example.com/",
					},
					RefreshTokenRotationEnabled: true,
				}
				r := protocol.TokenRequest{}
				r["grant_type"] = []string{"refresh_token"}
				r["client_id"] = []string{"app-id"}
				r["refresh_token"] = []string{"asdf"}
				refreshTokenHash := "hash1"
				offlineGrant := &oauth.OfflineGrant{
					ID:              "offline-grant-id",
					Attrs:           *session.NewAttrs("user-id"),
					InitialClientID: "app-id",
					RefreshTokens: []oauth.OfflineGrantRefreshToken{{
						ClientID:         "app-id",
						Scopes:           []string{"openid"},
						InitialTokenHash: refreshTokenHash,
					}},
					ExpireAtForResolvedSession: time.Date(2020, 02, 01, 1, 0, 0, 0, time.UTC),
				}
				rateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
				tokenService.EXPECT().ParseRefreshToken(gomock.Any(), "asdf").Return(&oauth.Authorization{}, offlineGrant, refreshTokenHash, nil)

				idTokenIssuer.EXPECT().PrepareIDToken(gomock.Any(), gomock.Any()).Return(&oidc.PrepareIDTokenResult{}, nil)
				idTokenIssuer.EXPECT().MakeIDTokenFromPreparationResult(gomock.Any(), gomock.Any()).Return("id-token", nil)

				tokenService.EXPECT().PrepareUserAccessGrantByRefreshToken(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, options handler.PrepareUserAccessGrantByRefreshTokenOptions) (*handler.PrepareUserAccessGrantByRefreshTokenResult, error) {
						result := &handler.PrepareUserAccessGrantByRefreshTokenResult{
							// The value is unimportant.
							PreparationResult: nil,
						}
						if options.ShouldRotateRefreshToken {
							result.RotateRefreshTokenResult = &oauth.RotateRefreshTokenResult{
								GrantID: "grant-id",
								Token:   "new-refresh-token",
							}
						}

						return result, nil
					})
				accessTokenEncoding.EXPECT().MakeUserAccessTokenFromPreparationResult(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, options oauth.MakeUserAccessTokenFromPreparationOptions) (*oauth.IssueAccessGrantResult, error) {
					return &oauth.IssueAccessGrantResult{
						Token:     "access-token",
						TokenType: "Bearer",
						ExpiresIn: 300,
					}, nil
				})

				event := access.NewEvent(clock.NowUTC(), "1.2.3.4", "UA")
				offlineGrantService.EXPECT().AccessOfflineGrant(gomock.Any(), "offline-grant-id", refreshTokenHash, &event, offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				offlineGrants.EXPECT().UpdateOfflineGrantDeviceInfo(gomock.Any(), "offline-grant-id", gomock.Any(), offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				ctx := context.Background()
				res := handle(ctx, req, r)
				So(res.Result().StatusCode, ShouldEqual, 200)
				var body map[string]interface{}
				err := json.Unmarshal(res.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["refresh_token"], ShouldEqual, "grant-id.new-refresh-token")
			})

			Convey("should not rotate refresh token if disabled", func() {
				req, _ := http.NewRequest("POST", "/token", nil)
				clientResolver.ClientConfigs["app-id"] = &config.OAuthClientConfig{
					ClientID: "app-id",
					RedirectURIs: []string{
						"https://example.com/",
					},
					RefreshTokenRotationEnabled: false,
				}
				r := protocol.TokenRequest{}
				r["grant_type"] = []string{"refresh_token"}
				r["client_id"] = []string{"app-id"}
				r["refresh_token"] = []string{"asdf"}
				refreshTokenHash := "hash1"
				offlineGrant := &oauth.OfflineGrant{
					ID:              "offline-grant-id",
					Attrs:           *session.NewAttrs("user-id"),
					InitialClientID: "app-id",
					RefreshTokens: []oauth.OfflineGrantRefreshToken{{
						ClientID:         "app-id",
						Scopes:           []string{"openid"},
						InitialTokenHash: refreshTokenHash,
					}},
					ExpireAtForResolvedSession: time.Date(2020, 02, 01, 1, 0, 0, 0, time.UTC),
				}
				rateLimiter.EXPECT().Allow(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
				tokenService.EXPECT().ParseRefreshToken(gomock.Any(), "asdf").Return(&oauth.Authorization{}, offlineGrant, refreshTokenHash, nil)

				idTokenIssuer.EXPECT().PrepareIDToken(gomock.Any(), gomock.Any()).Return(&oidc.PrepareIDTokenResult{}, nil)
				idTokenIssuer.EXPECT().MakeIDTokenFromPreparationResult(gomock.Any(), gomock.Any()).Return("id-token", nil)

				tokenService.EXPECT().PrepareUserAccessGrantByRefreshToken(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, options handler.PrepareUserAccessGrantByRefreshTokenOptions) (*handler.PrepareUserAccessGrantByRefreshTokenResult, error) {
						result := &handler.PrepareUserAccessGrantByRefreshTokenResult{
							// The value is unimportant.
							PreparationResult: nil,
						}
						if options.ShouldRotateRefreshToken {
							result.RotateRefreshTokenResult = &oauth.RotateRefreshTokenResult{
								Token: "new-refresh-token",
							}
						}

						return result, nil
					})
				accessTokenEncoding.EXPECT().MakeUserAccessTokenFromPreparationResult(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, options oauth.MakeUserAccessTokenFromPreparationOptions) (*oauth.IssueAccessGrantResult, error) {
					return &oauth.IssueAccessGrantResult{
						Token:     "access-token",
						TokenType: "Bearer",
						ExpiresIn: 300,
					}, nil
				})

				event := access.NewEvent(clock.NowUTC(), "1.2.3.4", "UA")
				offlineGrantService.EXPECT().AccessOfflineGrant(gomock.Any(), "offline-grant-id", refreshTokenHash, &event, offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				offlineGrants.EXPECT().UpdateOfflineGrantDeviceInfo(gomock.Any(), "offline-grant-id", gomock.Any(), offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				ctx := context.Background()
				res := handle(ctx, req, r)
				So(res.Result().StatusCode, ShouldEqual, 200)
				var body map[string]interface{}
				err := json.Unmarshal(res.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body, ShouldNotContainKey, "refresh_token")
			})

			Convey("rate limited", func() {
				req, _ := http.NewRequest("POST", "/token", nil)
				clientResolver.ClientConfigs["app-id"] = &config.OAuthClientConfig{
					ClientID: "app-id",
				}
				r := protocol.TokenRequest{}
				r["grant_type"] = []string{"refresh_token"}
				r["client_id"] = []string{"app-id"}
				r["refresh_token"] = []string{"asdf"}

				tokenService.EXPECT().ParseRefreshToken(gomock.Any(), "asdf").Times(0)

				spec := ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}
				rateLimiter.EXPECT().Allow(gomock.Any(), spec).Return(ratelimit.NewFailedReservation(spec), nil)

				ctx := context.Background()
				res := handle(ctx, req, r)
				So(res.Result().StatusCode, ShouldEqual, 429)

				var body map[string]interface{}
				err := json.Unmarshal(res.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["error"], ShouldEqual, "x_rate_limited")
				So(body["error_description"], ShouldEqual, "rate limit exceeded, please try again later.")
			})
		})

		Convey("token exchange: pre-authenticated-url-token", func() {
			req, _ := http.NewRequest("POST", "/token", nil)
			clientID1 := "client-id-1"
			clientID2 := "client-id-2"
			clientResolver.ClientConfigs[clientID1] = &config.OAuthClientConfig{
				ClientID: clientID1,
				RedirectURIs: []string{
					"https://example.com/",
				},
			}
			// nolint:gosec
			testDeviceSecret := "TEST_DEVICE_SECRET"
			testOfflineGrantID := "TEST_SESSION_ID"
			testIdToken := "TEST_ID_TOKEN"
			testUserId := "TEST_USER_ID"
			testScopes := []string{
				"openid",
				oauth.OfflineAccess,
				oauth.DeviceSSOScope,
				oauth.PreAuthenticatedURLScope,
			}
			dsHash := oauth.HashToken(testDeviceSecret)
			testOfflineGrant := &oauth.OfflineGrant{
				AppID:            appID,
				ID:               testOfflineGrantID,
				Attrs:            *session.NewAttrs(testUserId),
				InitialClientID:  clientID2,
				DeviceSecretHash: dsHash,
				RefreshTokens: []oauth.OfflineGrantRefreshToken{{
					ClientID: clientID2,
					Scopes:   testScopes,
				}},
			}
			rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
				Name:           ratelimit.OAuthTokenPerIP,
				RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
				RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
				Arguments:      []string{"1.2.3.4"},
				Period:         time.Minute,
				Burst:          120,
				Enabled:        true,
			}).Return(nil, nil)
			rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
				Name:           ratelimit.OAuthTokenPerUser,
				RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerUser,
				RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
				Arguments:      []string{testUserId},
				Period:         time.Minute,
				Burst:          60,
				Enabled:        true,
			}).Return(nil, nil)
			offlineGrantService.EXPECT().GetOfflineGrant(gomock.Any(), testOfflineGrantID).
				AnyTimes().
				Return(testOfflineGrant, nil)
			sid := oauth.EncodeSID(testOfflineGrant)
			mockIdToken := jwt.New()
			_ = mockIdToken.Set("iss", origin)
			_ = mockIdToken.Set("sid", sid)
			_ = mockIdToken.Set("ds_hash", dsHash)
			idTokenIssuer.EXPECT().VerifyIDToken(testIdToken).
				Return(mockIdToken, nil).
				Times(1)

			testAuthz := &oauth.Authorization{
				ClientID: clientID1,
				UserID:   testUserId,
				Scopes:   testScopes,
			}
			authorizations.EXPECT().CheckAndGrant(gomock.Any(), clientID1, testUserId, gomock.InAnyOrder(testScopes)).
				AnyTimes().
				Return(testAuthz, nil)

			// nolint:gosec
			expectedPreAuthenticatedURLToken := "TEST_PRE_AUTHENTICATED_URL_TOKEN"
			expectedPreAuthenticatedURLTokenHash := oauth.HashToken(expectedPreAuthenticatedURLToken)
			expectedPreAuthenticatedURLTokenType := "Bearer"
			expectedPreAuthenticatedURLTokenExpiresIn := 1234

			issuePreAuthenticatedURLTokenResult := &handler.IssuePreAuthenticatedURLTokenResult{
				Token:     expectedPreAuthenticatedURLToken,
				TokenHash: expectedPreAuthenticatedURLTokenHash,
				TokenType: expectedPreAuthenticatedURLTokenType,
				ExpiresIn: expectedPreAuthenticatedURLTokenExpiresIn,
			}
			preAuthenticatedURLService.EXPECT().
				// TODO: Implement a stricter matcher
				IssuePreAuthenticatedURLToken(gomock.Any(), gomock.AssignableToTypeOf((*handler.IssuePreAuthenticatedURLTokenOptions)(nil))).
				Times(1).
				Return(issuePreAuthenticatedURLTokenResult, nil)

			offlineGrants.EXPECT().UpdateOfflineGrantDeviceSecretHash(gomock.Any(), testOfflineGrantID, gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(testOfflineGrant, nil)

			expectedNewIdToken := "NEW_ID_TOKEN"
			idTokenIssuer.EXPECT().PrepareIDToken(gomock.Any(), gomock.Any()).Times(1).Return(&oidc.PrepareIDTokenResult{}, nil)
			idTokenIssuer.EXPECT().MakeIDTokenFromPreparationResult(gomock.Any(), gomock.Any()).Times(1).Return(expectedNewIdToken, nil)

			newDeviceSecret := "newdevicesecret"
			tokenService.EXPECT().IssueDeviceSecret(gomock.Any(), gomock.Any()).Times(1).Return("newdshash").Do(func(ctx context.Context, resp protocol.TokenResponse) {
				resp.DeviceSecret(newDeviceSecret)
			})

			request := protocol.TokenRequest{
				"client_id":            []string{clientID1},
				"grant_type":           []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
				"requested_token_type": []string{"urn:authgear:params:oauth:token-type:pre-authenticated-url-token"},
				"audience":             []string{"http://accounts.example.com"},
				"subject_token_type":   []string{"urn:ietf:params:oauth:token-type:id_token"},
				"subject_token":        []string{testIdToken},
				"actor_token_type":     []string{"urn:x-oath:params:oauth:token-type:device-secret"},
				"actor_token":          []string{testDeviceSecret},
			}
			ctx := context.Background()
			resp := handle(ctx, req, request)

			So(resp.Result().StatusCode, ShouldEqual, 200)
			var body map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &body)
			So(err, ShouldBeNil)

			So(body["access_token"], ShouldEqual, expectedPreAuthenticatedURLToken)
			So(body["device_secret"], ShouldEqual, newDeviceSecret)
			So(body["expires_in"], ShouldEqual, expectedPreAuthenticatedURLTokenExpiresIn)
			So(body["id_token"], ShouldEqual, expectedNewIdToken)
			So(body["issued_token_type"], ShouldEqual, "urn:authgear:params:oauth:token-type:pre-authenticated-url-token")
			So(body["token_type"], ShouldEqual, expectedPreAuthenticatedURLTokenType)
		})

		Convey("client_credentials flow", func() {
			clientID := "client-cred-client"
			resourceURI := "https://api.example.com/resource"
			resourceID := "resource-id-1"
			allowedScopes := []*resourcescope.Scope{
				{ID: "scope-id-1", ResourceID: resourceID, Scope: "read"},
				{ID: "scope-id-2", ResourceID: resourceID, Scope: "write"},
			}
			clientResolver.ClientConfigs[clientID] = &config.OAuthClientConfig{
				ClientID:            clientID,
				ApplicationType:     config.OAuthClientApplicationTypeConfidential,
				AccessTokenLifetime: config.DurationSeconds(3600),
				IssueJWTAccessToken: true,
			}

			// Mock the client secret for the client
			key, err := jwk.FromRaw([]byte("supersecret"))
			if err != nil {
				t.Fatalf("failed to create jwk: %v", err)
			}
			keySet := jwk.NewSet()
			_ = keySet.AddKey(key)
			h.OAuthClientCredentials = &config.OAuthClientCredentials{
				Items: []config.OAuthClientCredentialsItem{
					{
						ClientID:                     clientID,
						OAuthClientCredentialsKeySet: config.OAuthClientCredentialsKeySet{Set: keySet},
					},
				},
			}

			resource := &resourcescope.Resource{
				ID:          resourceID,
				ResourceURI: resourceURI,
			}
			Convey("success", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				clientResourceScopeService.EXPECT().GetClientResourceByURI(gomock.Any(), clientID, resourceURI).Return(resource, nil)
				clientResourceScopeService.EXPECT().GetClientResourceScopes(gomock.Any(), clientID, resourceID).Return(allowedScopes, nil)

				accessToken := "access-token-123"
				tokenService.EXPECT().IssueClientCredentialsAccessToken(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, opts handler.ClientCredentialsAccessTokenOptions, resp protocol.TokenResponse) error {
						resp.AccessToken(accessToken)
						resp.TokenType("Bearer")
						resp.ExpiresIn(3600)
						resp.Scope("read write")
						return nil
					},
				)

				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{resourceURI},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 200)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["access_token"], ShouldEqual, accessToken)
				So(body["token_type"], ShouldEqual, "Bearer")
				So(body["expires_in"], ShouldEqual, 3600)
				So(body["scope"], ShouldEqual, "read write")
			})

			Convey("request for subset of scopes", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				clientResourceScopeService.EXPECT().GetClientResourceByURI(gomock.Any(), clientID, resourceURI).Return(resource, nil)
				clientResourceScopeService.EXPECT().GetClientResourceScopes(gomock.Any(), clientID, resourceID).Return(allowedScopes, nil)

				accessToken := "access-token-123"
				tokenService.EXPECT().IssueClientCredentialsAccessToken(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, opts handler.ClientCredentialsAccessTokenOptions, resp protocol.TokenResponse) error {
						resp.AccessToken(accessToken)
						resp.TokenType("Bearer")
						resp.ExpiresIn(3600)
						resp.Scope(strings.Join(opts.Scopes, " "))
						return nil
					},
				)

				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{resourceURI},
					"scope":         []string{"read"},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 200)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["access_token"], ShouldEqual, accessToken)
				So(body["token_type"], ShouldEqual, "Bearer")
				So(body["expires_in"], ShouldEqual, 3600)
				So(body["scope"], ShouldEqual, "read")
			})

			Convey("request for invalid scopes", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				clientResourceScopeService.EXPECT().GetClientResourceByURI(gomock.Any(), clientID, resourceURI).Return(resource, nil)
				clientResourceScopeService.EXPECT().GetClientResourceScopes(gomock.Any(), clientID, resourceID).Return(allowedScopes, nil)

				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{resourceURI},
					"scope":         []string{"admin"},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 400)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["error"], ShouldEqual, "invalid_scope")
				So(body["error_description"], ShouldEqual, "specified scope is not allowed: admin")
			})

			Convey("request for invalid resource", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				clientResourceScopeService.EXPECT().GetClientResourceByURI(gomock.Any(), clientID, resourceURI).Return(nil, resourcescope.ErrResourceNotFound)

				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{resourceURI},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 400)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["error"], ShouldEqual, "invalid_target")
				So(body["error_description"], ShouldEqual, "resource not found")
			})

			Convey("request for resource not associated with client", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				clientResourceScopeService.EXPECT().GetClientResourceByURI(gomock.Any(), clientID, resourceURI).Return(resource, resourcescope.ErrResourceNotAssociatedWithClient)

				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{resourceURI},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 400)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["error"], ShouldEqual, "invalid_target")
				So(body["error_description"], ShouldEqual, "client is not associated with the resource")
			})

			Convey("resource uri prefixed with public origin is blocked", func() {
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenPerIP,
					RateLimitName:  ratelimit.RateLimitOAuthTokenGeneralPerIP,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenGeneral,
					Arguments:      []string{"1.2.3.4"},
					Period:         time.Minute,
					Burst:          120,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerClient,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerClient,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Arguments:      []string{clientID},
					Period:         time.Minute,
					Burst:          5,
					Enabled:        true,
				}).Return(nil, nil)
				rateLimiter.EXPECT().Allow(gomock.Any(), ratelimit.BucketSpec{
					Name:           ratelimit.OAuthTokenClientCredentialsPerProject,
					RateLimitName:  ratelimit.RateLimitOAuthTokenClientCredentialsPerProject,
					RateLimitGroup: ratelimit.RateLimitGroupOAuthTokenClientCredentials,
					Period:         time.Minute,
					Burst:          20,
					Enabled:        true,
				}).Return(nil, nil)
				issuerResourceURI := origin + "/some-resource"
				req, _ := http.NewRequest("POST", "/token", nil)
				r := protocol.TokenRequest{
					"grant_type":    []string{"client_credentials"},
					"client_id":     []string{clientID},
					"client_secret": []string{"supersecret"},
					"resource":      []string{issuerResourceURI},
				}
				ctx := context.Background()
				resp := handle(ctx, req, r)

				So(resp.Result().StatusCode, ShouldEqual, 400)
				var body map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &body)
				So(err, ShouldBeNil)
				So(body["error"], ShouldEqual, "invalid_target")
				So(body["error_description"], ShouldEqual, "resource URI must not be a prefixed by authgear endpoint")
			})
		})
	})
}
