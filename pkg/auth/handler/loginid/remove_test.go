package loginid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestRemoveLoginIDHandler(t *testing.T) {
	Convey("Test RemoveLoginIDHandler", t, func() {
		h := &RemoveLoginIDHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			RemoveLoginIDRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		authContext := authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1")
		h.AuthContext = authContext
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			[]config.LoginIDKeyConfiguration{
				newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 1, 1),
				newLoginIDKeyConfig("username", config.LoginIDKeyType(metadata.Username), 0, 1),
			},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"principal-id-1": password.Principal{
					ID:         "principal-id-1",
					UserID:     "user-id-1",
					LoginIDKey: "email",
					LoginID:    "user1@example.com",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"email": "user1@example.com",
					},
				},
				"principal-id-2": password.Principal{
					ID:         "principal-id-2",
					UserID:     "user-id-1",
					LoginIDKey: "username",
					LoginID:    "user1",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"username": "user1",
					},
				},
				"principal-id-3": password.Principal{
					ID:         "principal-id-3",
					UserID:     "user-id-2",
					LoginIDKey: "username",
					LoginID:    "user2",
					Realm:      password.DefaultRealm,
					ClaimsValue: map[string]interface{}{
						"username": "user2",
					},
				},
			},
		)
		h.PasswordAuthProvider = passwordAuthProvider
		h.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider)
		sessionProvider := session.NewMockProvider()
		sessionProvider.Sessions["user-id-1-principal-id-2"] = auth.Session{
			ID:          "user-id-1-principal-id-2",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-2",
		}
		h.SessionProvider = sessionProvider
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("should fail if login ID does not exist", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "username", "value": "user"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "LoginIDNotFound",
					"message": "login ID does not exist",
					"code": 404
				}
			}`)
		})

		Convey("should fail if login ID does not belong to the user", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "username", "value": "user2"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "LoginIDNotFound",
					"message": "login ID does not exist",
					"code": 404
				}
			}`)
		})

		Convey("should fail if attempted to delete current identity", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "email", "value": "user1@example.com"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "CurrentIdentityBeingDeleted",
					"message": "must not delete current identity",
					"code": 400
				}
			}`)
		})

		Convey("should fail if there are not enough login ID", func() {
			authContext.UseUser("user-id-1", "principal-id-2")
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "email", "value": "user1@example.com"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid login ID",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "EntryAmount",
								"message": "not enough login IDs",
								"pointer": "",
								"details": { "key": "email", "gte": 1 }
							}
						]
					}
				}
			}`)
		})

		Convey("should remove login ID", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "username", "value": "user1"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			So(passwordAuthProvider.PrincipalMap, ShouldHaveLength, 2)
			var p password.Principal
			err := passwordAuthProvider.GetPrincipalByLoginIDWithRealm("username", "user1", password.DefaultRealm, &p)
			So(err, ShouldBeError, "principal not found")

			So(sessionProvider.Sessions, ShouldBeEmpty)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityDeleteEvent{
					User: model.User{
						ID:         "user-id-1",
						Verified:   false,
						Disabled:   false,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   "principal-id-2",
						Type: "password",
						Attributes: principal.Attributes{
							"login_id_key": "username",
							"login_id":     "user1",
						},
						Claims: principal.Claims{
							"username": "user1",
						},
					},
				},
			})
		})
	})
}
