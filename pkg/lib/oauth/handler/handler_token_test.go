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
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestTokenHandler(t *testing.T) {

	Convey("Token handler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clientResolver := &mockClientResolver{}
		clientID1 := "client-id-1"
		clientID2 := "client-id-2"
		clientResolver.ClientConfig = &config.OAuthClientConfig{
			ClientID: clientID1,
			RedirectURIs: []string{
				"https://example.com/",
			},
		}
		origin := "http://accounts.example.com"
		idTokenIssuer := NewMockIDTokenIssuer(ctrl)
		idTokenIssuer.EXPECT().Iss().Return(origin).AnyTimes()

		offlineGrants := NewMockOAuthOfflineGrantStore(ctrl)

		authorizations := NewMockAuthorizationService(ctrl)

		appInitiatedSSOToWebTokenService := NewMockAppInitiatedSSOToWebTokenService(ctrl)

		appID := "testapp"

		tokenService := handler.TokenService{
			GenerateToken: func() string { return "NEW_TOKEN" },
		}

		offlineGrantService := NewMockOfflineGrantService(ctrl)

		h := handler.TokenHandler{
			Context:                context.Background(),
			AppID:                  config.AppID(appID),
			AppDomains:             []string{},
			HTTPProto:              "http",
			HTTPOrigin:             httputil.HTTPOrigin(origin),
			OAuthFeatureConfig:     &config.OAuthFeatureConfig{},
			IdentityFeatureConfig:  &config.IdentityFeatureConfig{},
			OAuthClientCredentials: &config.OAuthClientCredentials{},
			Logger:                 handler.TokenHandlerLogger{logrus.NewEntry(logrus.New())},

			TokenService:                     tokenService,
			ClientResolver:                   clientResolver,
			Authorizations:                   authorizations,
			OfflineGrants:                    offlineGrants,
			OfflineGrantService:              offlineGrantService,
			IDTokenIssuer:                    idTokenIssuer,
			AppInitiatedSSOToWebTokenService: appInitiatedSSOToWebTokenService,
		}

		handle := func(r protocol.TokenRequest) *httptest.ResponseRecorder {
			result := h.Handle(&httptest.ResponseRecorder{}, &http.Request{}, r)
			req, _ := http.NewRequest("GET", "/authorize", nil)
			resp := httptest.NewRecorder()
			result.WriteResponse(resp, req)
			return resp
		}

		Convey("token exchange: app-initiated-sso-to-web-token", func() {
			// nolint:gosec
			testDeviceSecret := "TEST_DEVICE_SECRET"
			testOfflineGrantID := "TEST_SESSION_ID"
			testIdToken := "TEST_ID_TOKEN"
			testUserId := "TEST_USER_ID"
			testScopes := []string{
				"openid",
				oauth.OfflineAccess,
				oauth.DeviceSSOScope,
				oauth.AppInitiatedSSOToWebScope,
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
			offlineGrants.EXPECT().GetOfflineGrant(testOfflineGrantID).
				AnyTimes().
				Return(testOfflineGrant, nil)
			sid := oidc.EncodeSID(testOfflineGrant)
			mockIdToken := jwt.New()
			_ = mockIdToken.Set("iss", origin)
			_ = mockIdToken.Set("sid", sid)
			_ = mockIdToken.Set("ds_hash", dsHash)
			idTokenIssuer.EXPECT().VerifyIDTokenWithoutClient(testIdToken).
				Return(mockIdToken, nil).
				Times(1)

			testAuthz := &oauth.Authorization{
				ClientID: clientID1,
				UserID:   testUserId,
				Scopes:   testScopes,
			}
			authorizations.EXPECT().CheckAndGrant(clientID1, testUserId, gomock.InAnyOrder(testScopes)).
				AnyTimes().
				Return(testAuthz, nil)

			// nolint:gosec
			expectedAppInitiatedSSOToWebToken := "TEST_APP_INITIATED_SSO_TO_WEB_TOKEN"
			expectedAppInitiatedSSOToWebTokenHash := oauth.HashToken(expectedAppInitiatedSSOToWebToken)
			expectedAppInitiatedSSOToWebTokenType := "Bearer"
			expectedAppInitiatedSSOToWebTokenExpiresIn := 1234

			issueAppInitiatedSSOToWebTokenResult := &oauth.IssueAppInitiatedSSOToWebTokenResult{
				Token:     expectedAppInitiatedSSOToWebToken,
				TokenHash: expectedAppInitiatedSSOToWebTokenHash,
				TokenType: expectedAppInitiatedSSOToWebTokenType,
				ExpiresIn: expectedAppInitiatedSSOToWebTokenExpiresIn,
			}
			appInitiatedSSOToWebTokenService.EXPECT().
				// TODO: Implement a stricter matcher
				IssueAppInitiatedSSOToWebToken(gomock.AssignableToTypeOf((*oauth.IssueAppInitiatedSSOToWebTokenOptions)(nil))).
				Times(1).
				Return(issueAppInitiatedSSOToWebTokenResult, nil)

			offlineGrantService.EXPECT().ComputeOfflineGrantExpiry(gomock.Any()).
				AnyTimes().
				Return(time.Now(), nil)

			offlineGrants.EXPECT().UpdateOfflineGrantDeviceSecretHash(testOfflineGrantID, gomock.Any(), gomock.Any()).
				AnyTimes().
				Return(testOfflineGrant, nil)

			expectedNewIdToken := "NEW_ID_TOKEN"
			idTokenIssuer.EXPECT().IssueIDToken(gomock.Any()).Times(1).Return(expectedNewIdToken, nil)

			request := protocol.TokenRequest{
				"client_id":            clientID1,
				"grant_type":           "urn:ietf:params:oauth:grant-type:token-exchange",
				"requested_token_type": "urn:authgear:params:oauth:token-type:app-initiated-sso-to-web-token",
				"audience":             "http://accounts.example.com",
				"subject_token_type":   "urn:ietf:params:oauth:token-type:id_token",
				"subject_token":        testIdToken,
				"actor_token_type":     "urn:x-oath:params:oauth:token-type:device-secret",
				"actor_token":          testDeviceSecret,
			}
			resp := handle(request)

			So(resp.Result().StatusCode, ShouldEqual, 200)
			var body map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &body)
			So(err, ShouldBeNil)

			So(body["access_token"], ShouldEqual, expectedAppInitiatedSSOToWebToken)
			So(body["device_secret"], ShouldNotBeNil)
			So(body["expires_in"], ShouldEqual, expectedAppInitiatedSSOToWebTokenExpiresIn)
			So(body["id_token"], ShouldEqual, expectedNewIdToken)
			So(body["issued_token_type"], ShouldEqual, "urn:authgear:params:oauth:token-type:app-initiated-sso-to-web-token")
			So(body["token_type"], ShouldEqual, expectedAppInitiatedSSOToWebTokenType)
		})
	})
}
