package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestRevokeHandler(t *testing.T) {
	Convey("Test RevokeHandler", t, func() {
		h := &RevokeHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			RevokeRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		authContext := authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1")
		h.AuthContext = authContext
		sessionProvider := session.NewMockProvider()
		h.SessionProvider = sessionProvider
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			[]config.LoginIDKeyConfiguration{},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"principal-id-2": password.Principal{
					ID:         "principal-id-2",
					UserID:     "user-id-1",
					LoginIDKey: "email",
					LoginID:    "user@example.com",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"email": "user@example.com",
					},
				},
			},
		)
		h.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		now := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
		sessionProvider.Sessions["user-id-1-principal-id-1"] = auth.Session{
			ID:          "user-id-1-principal-id-1",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-1",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		sessionProvider.Sessions["user-id-1-principal-id-2"] = auth.Session{
			ID:          "user-id-1-principal-id-2",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-2",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		sessionProvider.Sessions["user-id-2-principal-id-3"] = auth.Session{
			ID:          "user-id-2-principal-id-3",
			ClientID:    "web-app",
			UserID:      "user-id-2",
			PrincipalID: "principal-id-3",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		sess := sessionProvider.Sessions["user-id-1-principal-id-1"]
		authContext.UseSession(&sess)

		Convey("should revoke existing session", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-1-principal-id-2"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, struct{}{})

			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldNotContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.SessionDeleteEvent{
					Reason: event.SessionDeleteReasonRevoke,
					User: model.User{
						ID:         "user-id-1",
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   "principal-id-2",
						Type: "password",
						Attributes: principal.Attributes{
							"login_id_key": "email",
							"login_id":     "user@example.com",
						},
						Claims: principal.Claims{
							"email": "user@example.com",
						},
					},
					Session: model.Session{
						ID:             "user-id-1-principal-id-2",
						IdentityID:     "principal-id-2",
						CreatedAt:      now,
						LastAccessedAt: now,
					},
				},
			})
		})

		Convey("should reject current session", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"session_id": "user-id-1-principal-id-1"
			}`))
			r.Header.Set("Content-Type", "application/json")
			r = auth.InitRequestAuthContext(r)
			authContext.CopyTo(auth.NewContextSetterWithContext(r.Context()))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "General", "message": "session_id must not be current session", "pointer": "/session_id" }
						]
					}
				}
			}`)

			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")
		})

		Convey("should ignore non-existing session", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-1-principal-id-4"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
		})

		Convey("should ignore session of other users", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-2-principal-id-3"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")
		})

	})
}
