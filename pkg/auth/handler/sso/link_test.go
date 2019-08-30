package sso

import (
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
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	coreconfig "github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLinkPayload(t *testing.T) {
	Convey("Test LinkRequestPayload", t, func() {
		// callback URL and ux_mode is required
		Convey("validate valid payload", func() {
			payload := LinkRequestPayload{
				AccessToken: "token",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without access token", func() {
			payload := LinkRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}

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
		sh.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
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
			UserInfo:       providerUserInfo,
		}
		sh.Provider = &mockProvider
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
		sh.OAuthConfiguration = oauthConfig
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should link user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
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
