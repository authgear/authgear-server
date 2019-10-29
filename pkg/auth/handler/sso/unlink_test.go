package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerID := "google"
		providerUserID := "mock_user_id"

		req, _ := http.NewRequest("POST", "https://api.example.com", nil)

		sh := &UnlinkHandler{}
		sh.ProviderID = providerID
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
		sh.ProviderFactory = sso.NewProviderFactory(config.TenantConfiguration{
			UserConfig: config.UserConfiguration{
				SSO: config.SSOConfiguration{
					OAuth: config.OAuthConfiguration{
						Providers: []config.OAuthProviderConfiguration{
							config.OAuthProviderConfiguration{
								Type: "google",
								ID:   "google",
							},
						},
					},
				},
			},
		}, urlprefix.NewProvider(req))
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
		sh.IdentityProvider = principal.NewMockIdentityProvider(mockOAuthProvider)
		sh.OAuthAuthProvider = mockOAuthProvider
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		sessionProvider := session.NewMockProvider()
		sh.SessionProvider = sessionProvider
		sessionProvider.Sessions["faseng.cat.id-faseng.cat.principal.id"] = auth.Session{
			ID:          "faseng.cat.id-faseng.cat.principal.id",
			UserID:      "faseng.cat.id",
			PrincipalID: "faseng.cat.principal.id",
		}
		sessionProvider.Sessions["faseng.cat.id-oauth-principal-id"] = auth.Session{
			ID:          "faseng.cat.id-oauth-principal-id",
			UserID:      "faseng.cat.id",
			PrincipalID: "oauth-principal-id",
		}

		Convey("should unlink user id with oauth principal", func() {
			h := handler.APIHandlerToHandler(sh, sh.TxContext)
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
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

			So(sessionProvider.Sessions, ShouldContainKey, "faseng.cat.id-faseng.cat.principal.id")
			So(sessionProvider.Sessions, ShouldNotContainKey, "faseng.cat.id-oauth-principal-id")

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{
						ID:         "faseng.cat.id",
						Verified:   true,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   "oauth-principal-id",
						Type: "oauth",
						Attributes: principal.Attributes{
							"provider_keys":    map[string]interface{}{},
							"provider_type":    "google",
							"provider_user_id": "mock_user_id",
							"raw_profile":      nil,
						},
						Claims: principal.Claims{
							"email": "faseng@example.com",
						},
					},
				},
			})
		})

		Convey("should disallow remove current identity", func() {
			sh.AuthContext = authtest.NewMockContext().
				UseUser("faseng.cat.id", "oauth-principal-id").
				MarkVerified()
			h := handler.APIHandlerToHandler(sh, sh.TxContext)

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "CurrentIdentityBeingDeleted",
					"message": "must not delete current identity",
					"code": 400
				}
			}`)
		})
	})
}
