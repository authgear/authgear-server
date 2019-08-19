package sso

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoginPayload(t *testing.T) {
	Convey("Test LoginRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := LoginRequestPayload{
				AccessToken:     "token",
				OnUserDuplicate: model.OnUserDuplicateDefault,
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without access token", func() {
			payload := LoginRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

func TestLoginHandler(t *testing.T) {
	realTime := timeNow
	now := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	timeNow = func() time.Time { return now }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test LoginHandler", t, func() {
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserID := "mock_user_id"

		sh := &LoginHandler{}
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		oauthConfig := coreconfig.OAuthConfiguration{
			URLPrefix:                      "http://localhost:3000",
			StateJWTSecret:                 stateJWTSecret,
			ExternalAccessTokenFlowEnabled: true,
			AllowedCallbackURLs: []string{
				"http://localhost",
			},
		}
		providerConfig := coreconfig.OAuthProviderConfiguration{
			ID:           providerName,
			Type:         "google",
			ClientID:     "mock_client_id",
			ClientSecret: "mock_client_secret",
		}
		mockProvider := sso.MockSSOProvider{
			BaseURL:        "http://mock/auth",
			OAuthConfig:    oauthConfig,
			ProviderConfig: providerConfig,
			UserInfo: sso.ProviderUserInfo{
				ID:    providerUserID,
				Email: "mock@example.com",
			},
		}
		sh.Provider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(nil)
		sh.OAuthAuthProvider = mockOAuthProvider
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider)
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.OAuthConfiguration = oauthConfig
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should get auth response", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			token := mockTokenStore.GetTokensByAuthInfoID(p.UserID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "%s",
						"type": "oauth",
						"provider_type": "google",
						"provider_user_id": "mock_user_id",
						"raw_profile": {
							"id": "mock_user_id",
							"email": "mock@example.com"
						},
						"claims": {
							"email": "mock@example.com"
						}
					},
					"access_token": "%s"
				}
			}`,
				p.UserID,
				p.ID,
				token.AccessToken))

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserCreateEvent{
					User: model.User{
						ID:          p.UserID,
						LastLoginAt: &now,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identities: []model.Identity{
						model.Identity{
							ID:   p.ID,
							Type: "oauth",
							Attributes: principal.Attributes{
								"provider_type":    "google",
								"provider_user_id": "mock_user_id",
								"raw_profile": map[string]interface{}{
									"id":    "mock_user_id",
									"email": "mock@example.com",
								},
							},
							Claims: principal.Claims{
								"email": "mock@example.com",
							},
						},
					},
				},
				event.SessionCreateEvent{
					Reason: event.SessionCreateReasonSignup,
					User: model.User{
						ID:          p.UserID,
						LastLoginAt: &now,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   p.ID,
						Type: "oauth",
						Attributes: principal.Attributes{
							"provider_type":    "google",
							"provider_user_id": "mock_user_id",
							"raw_profile": map[string]interface{}{
								"id":    "mock_user_id",
								"email": "mock@example.com",
							},
						},
						Claims: principal.Claims{
							"email": "mock@example.com",
						},
					},
				},
			})
		})

		sh.OAuthConfiguration.ExternalAccessTokenFlowEnabled = false
		h = handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should return error if disabled", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "External access token flow is disabled",
					"name": "UndefinedOperation"
				}
			}`)
		})
	})
}
