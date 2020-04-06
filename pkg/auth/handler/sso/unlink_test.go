package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
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

type mockRemoveSessionManager struct {
	Sessions []auth.AuthSession
}

func (m *mockRemoveSessionManager) List(userID string) ([]auth.AuthSession, error) {
	return m.Sessions, nil
}

func (m *mockRemoveSessionManager) Revoke(s auth.AuthSession) error {
	n := 0
	for _, session := range m.Sessions {
		if session.SessionID() == s.SessionID() {
			continue
		}
		m.Sessions[n] = session
		n++
	}
	m.Sessions = m.Sessions[:n]
	return nil
}

type mockUnlinkSessionManager struct {
	Sessions []auth.AuthSession
}

func (m *mockUnlinkSessionManager) List(userID string) ([]auth.AuthSession, error) {
	return m.Sessions, nil
}

func (m *mockUnlinkSessionManager) Revoke(s auth.AuthSession) error {
	n := 0
	for _, session := range m.Sessions {
		if session.SessionID() == s.SessionID() {
			continue
		}
		m.Sessions[n] = session
		n++
	}
	m.Sessions = m.Sessions[:n]
	return nil
}

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerID := "google"
		providerUserID := "mock_user_id"

		req, _ := http.NewRequest("POST", "https://api.example.com", nil)

		sh := &UnlinkHandler{}
		sh.ProviderID = providerID
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
		}, urlprefix.NewProvider(req), timeProvider, nil, nil)
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
		sessionManager := &mockRemoveSessionManager{}
		sessionManager.Sessions = []auth.AuthSession{
			authtesting.WithAuthn().
				SessionID("faseng.cat.id-faseng.cat.principal.id").
				UserID("faseng.cat.id").
				PrincipalID("faseng.cat.principal.id").
				ToSession(),
			authtesting.WithAuthn().
				SessionID("faseng.cat.id-faseng.oauth-principal.id").
				UserID("faseng.cat.id").
				PrincipalID("oauth-principal-id").
				ToSession(),
		}
		sh.SessionManager = sessionManager

		Convey("should unlink user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				PrincipalID("faseng.cat.principal.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
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

			So(sessionManager.Sessions, ShouldHaveLength, 1)
			So(sessionManager.Sessions[0].SessionID(), ShouldEqual, "faseng.cat.id-faseng.cat.principal.id")

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{
						ID:         "faseng.cat.id",
						Verified:   false,
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
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				PrincipalID("oauth-principal-id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
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

		Convey("should error on unknown identity", func() {
			sh.OAuthAuthProvider = oauth.NewMockProvider(nil)
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
			}`))
			req = authtesting.WithAuthn().
				UserID("faseng.cat.id").
				ToRequest(req)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"code": 404,
					"message": "oauth principal not found",
					"name": "NotFound",
					"reason": "OAuthPrincipalNotFound"
				}
			}
			`)
		})
	})
}
