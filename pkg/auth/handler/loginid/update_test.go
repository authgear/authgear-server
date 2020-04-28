package loginid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	authtesting "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/testing"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type mockUpdateSessionManager struct {
	Sessions []auth.AuthSession
}

func (m *mockUpdateSessionManager) List(userID string) ([]auth.AuthSession, error) {
	return m.Sessions, nil
}

func (m *mockUpdateSessionManager) Update(s auth.AuthSession) error {
	for i, session := range m.Sessions {
		if session.SessionID() != s.SessionID() {
			continue
		}
		m.Sessions[i] = s
	}
	return nil
}

func TestUpdateLoginIDHandler(t *testing.T) {
	Convey("Test UpdateLoginIDHandler", t, func() {
		h := &UpdateLoginIDHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			UpdateLoginIDRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		authctx := authtesting.WithAuthn().
			UserID("user-id-1").
			Verified(true)
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"user-id-1": authinfo.AuthInfo{
					ID:         "user-id-1",
					Verified:   true,
					VerifyInfo: map[string]bool{"user1@example.com": true},
				},
			},
		)
		h.AuthInfoStore = authInfoStore
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			[]config.LoginIDKeyConfiguration{
				newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 1),
				newLoginIDKeyConfig("username", config.LoginIDKeyType(metadata.Username), 1),
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
		h.UserVerificationProvider = userverify.NewProvider(nil, nil, &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email"},
			},
		}, nil)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("should fail if login ID does not exist", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": { "key": "username", "value": "user" },
				"new_login_id": { "key": "username", "value": "user1_a" }
			}`))
			r = authctx.ToRequest(r)
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
				"old_login_id": { "key": "username", "value": "user2" },
				"new_login_id": { "key": "username", "value": "user1_a" }
			}`))
			r = authctx.ToRequest(r)
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

		Convey("should fail if login ID is already used", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": { "key": "username", "value": "user1" },
				"new_login_id": { "key": "username", "value": "user2" }
			}`))
			r = authctx.ToRequest(r)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is already used",
					"code": 409
				}
			}`)
		})

		Convey("should fail if login ID amount is invalid", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": { "key": "email", "value": "user1@example.com" },
				"new_login_id": { "key": "username", "value": "user1_a" }
			}`))
			r = authctx.ToRequest(r)
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
								"message": "too many login IDs",
								"pointer": "",
								"details": { "key": "username", "lte": 1 }
							}
						]
					}
				}
			}`)
		})

		Convey("should update login ID", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": { "key": "email", "value": "user1@example.com" },
				"new_login_id": { "key": "email", "value": "user1+a@example.com" }
			}`))
			r = authctx.ToRequest(r)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 200)

			So(passwordAuthProvider.PrincipalMap, ShouldHaveLength, 3)
			_, err := passwordAuthProvider.GetPrincipalByLoginID("email", "user1@example.com")
			So(err, ShouldBeError, "principal not found")
			p, err := passwordAuthProvider.GetPrincipalByLoginID("email", "user1+a@example.com")
			So(err, ShouldBeNil)
			So(p.UserID, ShouldEqual, "user-id-1")
			So(p.LoginIDKey, ShouldEqual, "email")
			So(p.LoginID, ShouldEqual, "user1+a@example.com")

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user": {
						"id": "user-id-1",
						"created_at": "0001-01-01T00:00:00Z",
						"is_disabled": false,
						"is_manually_verified": false,
						"is_verified": true,
						"verify_info": {"user1@example.com":true},
						"metadata": {}
					},
					"identity": {
						"type": "password",
						"claims": {
							"email": "user1+a@example.com",
							"https://auth.skygear.io/claims/login_id/key":   "email",
							"https://auth.skygear.io/claims/login_id/value": "user1+a@example.com"
						}
					}
				}
			}`)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.IdentityCreateEvent{
					User: model.User{
						ID:         "user-id-1",
						Verified:   true,
						Disabled:   false,
						VerifyInfo: map[string]bool{"user1@example.com": true},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						Type: "password",
						Claims: principal.Claims{
							"email": "user1+a@example.com",
							"https://auth.skygear.io/claims/login_id/key":   "email",
							"https://auth.skygear.io/claims/login_id/value": "user1+a@example.com",
						},
					},
				},
				event.IdentityDeleteEvent{
					User: model.User{
						ID:         "user-id-1",
						Verified:   true,
						Disabled:   false,
						VerifyInfo: map[string]bool{"user1@example.com": true},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						Type: "password",
						Claims: principal.Claims{
							"email": "user1@example.com",
							"https://auth.skygear.io/claims/login_id/key":   "email",
							"https://auth.skygear.io/claims/login_id/value": "user1@example.com",
						},
					},
				},
			})
		})

		SkipConvey("should invalidate verify state", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": { "key": "email", "value": "user1@example.com" },
				"new_login_id": { "key": "email", "value": "user1+a@example.com" }
			}`))
			r = authctx.ToRequest(r)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 200)

			So(authInfoStore.AuthInfoMap["user-id-1"].VerifyInfo["user1@example.com"], ShouldBeFalse)
			So(authInfoStore.AuthInfoMap["user-id-1"].Verified, ShouldBeFalse)
		})
	})
}
