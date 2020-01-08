package sso

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLinkHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test LinkHandler", t, func() {
		stateJWTSecret := "secret"
		providerName := "mock"
		providerUserInfo := sso.ProviderUserInfo{
			ID:    "mock_user_id",
			Email: "john.doe@example.com",
		}

		sh := &LinkHandler{}
		sh.TxContext = db.NewMockTxContext()
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			LinkRequestSchema,
		)
		sh.Validator = validator
		sh.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
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
			UserInfo:       providerUserInfo,
		}
		sh.OAuthProvider = &mockProvider
		sh.SSOProvider = &mockProvider
		mockOAuthProvider := oauth.NewMockProvider(nil)
		sh.OAuthAuthProvider = mockOAuthProvider
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.OAuthAuthProvider)
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"faseng.cat.id": authinfo.AuthInfo{
					ID: "faseng.cat.id",
				},
			},
		)
		sh.AuthInfoStore = authInfoStore
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider

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

		Convey("should link user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "faseng.cat.id",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": false,
						"metadata": {},
						"verify_info": {}
					}
				}
			}`)

			p, _ := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserInfo.ID,
			})
			So(p.UserID, ShouldEqual, "faseng.cat.id")

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityCreateEvent{
					User: model.User{
						ID:         "faseng.cat.id",
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
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
								"email": "john.doe@example.com",
							},
						},
						Claims: principal.Claims{
							"email": "john.doe@example.com",
						},
					},
				},
			})
		})

		mockProvider.OAuthConfig.ExternalAccessTokenFlowEnabled = false

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
