package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestTokenHandler(t *testing.T) {

	Convey("Token handler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clientResolver := &mockClientResolver{}
		origin := "http://accounts.example.com"
		idTokenIssuer := NewMockIDTokenIssuer(ctrl)
		idTokenIssuer.EXPECT().Iss().Return(origin).AnyTimes()

		offlineGrants := NewMockTokenHandlerOfflineGrantStore(ctrl)

		authorizations := NewMockAuthorizationService(ctrl)

		preAuthenticatedURLService := NewMockPreAuthenticatedURLTokenService(ctrl)

		appID := "testapp"

		tokenService := NewMockTokenHandlerTokenService(ctrl)

		offlineGrantService := NewMockTokenHandlerOfflineGrantService(ctrl)

		clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")

		h := handler.TokenHandler{
			AppID:                  config.AppID(appID),
			AppDomains:             []string{},
			HTTPProto:              "http",
			HTTPOrigin:             httputil.HTTPOrigin(origin),
			OAuthFeatureConfig:     &config.OAuthFeatureConfig{},
			IdentityFeatureConfig:  &config.IdentityFeatureConfig{},
			OAuthClientCredentials: &config.OAuthClientCredentials{},
			Logger:                 handler.TokenHandlerLogger{logrus.NewEntry(logrus.New())},
			Clock:                  clock,
			RemoteIP:               "1.2.3.4",
			UserAgentString:        "UA",

			TokenService:                    tokenService,
			ClientResolver:                  clientResolver,
			Authorizations:                  authorizations,
			OfflineGrants:                   offlineGrants,
			OfflineGrantService:             offlineGrantService,
			IDTokenIssuer:                   idTokenIssuer,
			PreAuthenticatedURLTokenService: preAuthenticatedURLService,
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
				clientResolver.ClientConfig = &config.OAuthClientConfig{
					ClientID: "app-id",
					RedirectURIs: []string{
						"https://example.com/",
					},
				}
				r := protocol.TokenRequest{}
				r["grant_type"] = "refresh_token"
				r["client_id"] = "app-id"
				r["refresh_token"] = "asdf"
				refreshTokenHash := "hash1"
				offlineGrant := &oauth.OfflineGrant{
					ID:              "offline-grant-id",
					InitialClientID: "app-id",
					RefreshTokens: []oauth.OfflineGrantRefreshToken{{
						ClientID:  "app-id",
						Scopes:    []string{"openid"},
						TokenHash: refreshTokenHash,
					}},
					ExpireAtForResolvedSession: time.Date(2020, 02, 01, 1, 0, 0, 0, time.UTC),
				}
				tokenService.EXPECT().ParseRefreshToken(gomock.Any(), "asdf").Return(&oauth.Authorization{}, offlineGrant, refreshTokenHash, nil)
				idTokenIssuer.EXPECT().IssueIDToken(gomock.Any(), gomock.Any()).Return("id-token", nil)
				tokenService.EXPECT().IssueAccessGrant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				event := access.NewEvent(clock.NowUTC(), "1.2.3.4", "UA")
				offlineGrantService.EXPECT().AccessOfflineGrant(gomock.Any(), "offline-grant-id", &event, offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				offlineGrants.EXPECT().UpdateOfflineGrantDeviceInfo(gomock.Any(), "offline-grant-id", gomock.Any(), offlineGrant.ExpireAtForResolvedSession).Return(offlineGrant, nil)
				ctx := context.Background()
				res := handle(ctx, req, r)
				So(res.Result().StatusCode, ShouldEqual, 200)
			})
		})

		Convey("token exchange: pre-authenticated-url-token", func() {
			req, _ := http.NewRequest("POST", "/token", nil)
			clientID1 := "client-id-1"
			clientID2 := "client-id-2"
			clientResolver.ClientConfig = &config.OAuthClientConfig{
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
			idTokenIssuer.EXPECT().IssueIDToken(gomock.Any(), gomock.Any()).Times(1).Return(expectedNewIdToken, nil)

			newDeviceSecret := "newdevicesecret"
			tokenService.EXPECT().IssueDeviceSecret(gomock.Any(), gomock.Any()).Times(1).Return("newdshash").Do(func(ctx context.Context, resp protocol.TokenResponse) {
				resp.DeviceSecret(newDeviceSecret)
			})

			request := protocol.TokenRequest{
				"client_id":            clientID1,
				"grant_type":           "urn:ietf:params:oauth:grant-type:token-exchange",
				"requested_token_type": "urn:authgear:params:oauth:token-type:pre-authenticated-url-token",
				"audience":             "http://accounts.example.com",
				"subject_token_type":   "urn:ietf:params:oauth:token-type:id_token",
				"subject_token":        testIdToken,
				"actor_token_type":     "urn:x-oath:params:oauth:token-type:device-secret",
				"actor_token":          testDeviceSecret,
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
	})
}
