package sso

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/validation"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

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
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			LoginRequestSchema,
		)
		sh.Validator = validator
		oauthConfig := &coreconfig.OAuthConfiguration{
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
			URLPrefix:      &url.URL{Scheme: "https", Host: "api.example.com"},
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
		identityProvider := principal.NewMockIdentityProvider(sh.OAuthAuthProvider)
		sh.IdentityProvider = identityProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{},
		)
		sh.AuthInfoStore = authInfoStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStore()
		sh.UserProfileStore = userProfileStore
		sh.OAuthConfiguration = oauthConfig
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := &coreconfig.MFAConfiguration{
			Enabled:     false,
			Enforcement: coreconfig.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		sh.AuthnSessionProvider = authnsession.NewMockProvider(
			mfaConfiguration,
			timeProvider,
			mfaProvider,
			authInfoStore,
			sessionProvider,
			sessionWriter,
			identityProvider,
			hookProvider,
			userProfileStore,
		)

		Convey("should reject payload without access token", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "Required",
								"message": "access_token is required",
								"pointer": "/access_token"
							}
						]
					}
				}
			}`)
		})

		Convey("should get auth response", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
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
						"provider_keys": {},
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
					"access_token": "access-token-%s-%s-0",
					"session_id": "%s-%s-0"
				}
			}`, p.UserID, p.ID, p.UserID, p.ID, p.UserID, p.ID))

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
								"provider_keys":    map[string]interface{}{},
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
					Reason: coreAuth.SessionCreateReasonSignup,
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
							"provider_keys":    map[string]interface{}{},
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
					Session: model.Session{
						ID:                fmt.Sprintf("%s-%s-0", p.UserID, p.ID),
						IdentityID:        p.ID,
						IdentityType:      "oauth",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
		})

		sh.OAuthConfiguration.ExternalAccessTokenFlowEnabled = false

		Convey("should return error if disabled", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "NotFound",
					"message": "external access token flow is disabled",
					"code": 404
				}
			}`)
		})
	})
}
