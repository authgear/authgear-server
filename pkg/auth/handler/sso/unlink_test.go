package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerUserID := "mock_user_id"

		req, _ := http.NewRequest("POST", "https://api.example.com", nil)

		sh := &UnlinkHandler{}
		sh.TxContext = db.NewMockTxContext()
		timeProvider := &coreTime.MockProvider{}
		sh.ProviderFactory = sso.NewOAuthProviderFactory(config.TenantConfiguration{
			AppConfig: &config.AppConfiguration{
				Identity: &config.IdentityConfiguration{
					OAuth: &config.OAuthConfiguration{
						Providers: []config.OAuthProviderConfiguration{
							config.OAuthProviderConfiguration{
								Type: "google",
								ID:   "google",
							},
						},
					},
				},
			},
		}, urlprefix.NewProvider(req), timeProvider, nil, nil, RedirectURIForAPI)
		mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:             "oauth-principal-id",
				ProviderType:   "google",
				ProviderKeys:   map[string]interface{}{},
				ProviderUserID: providerUserID,
				UserID:         "faseng.cat.id",
				ClaimsValue: map[string]interface{}{
					"email": "faseng@example.com",
				},
			},
		})
		sh.OAuthAuthProvider = mockOAuthProvider
		sh.AuthInfoStore = authinfo.NewMockStoreWithAuthInfoMap(map[string]authinfo.AuthInfo{
			"faseng.cat.id": {ID: "faseng.cat.id"},
		})
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider

		router := mux.NewRouter()
		router.Handle("/sso/{provider}/unlink", sh)

		Convey("should unlink user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "/sso/google/unlink", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			p, e := sh.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
				ProviderType:   "google",
				ProviderUserID: providerUserID,
			})
			So(e, ShouldBeError, principal.ErrNotFound)
			So(p, ShouldBeNil)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{
						ID:         "faseng.cat.id",
						Verified:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						Type: "oauth",
						Claims: principal.Claims{
							"email": "faseng@example.com",
							"https://auth.skygear.io/claims/oauth/provider": map[string]interface{}{
								"type": "google",
							},
							"https://auth.skygear.io/claims/oauth/subject_id": "mock_user_id",
							"https://auth.skygear.io/claims/oauth/profile":    nil,
						},
					},
				},
			})
		})

		Convey("should error on unknown identity", func() {
			sh.OAuthAuthProvider = oauth.NewMockProvider(nil)
			req, _ := http.NewRequest("POST", "/sso/google/unlink", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 404,
					"message": "oauth identity not found",
					"name": "NotFound",
					"reason": "OAuthIdentityNotFound"
				}
			}
			`)
		})
	})
}
