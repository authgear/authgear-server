package loginid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func newLoginIDKeyConfig(key string, t config.LoginIDKeyType, max int) config.LoginIDKeyConfiguration {
	return config.LoginIDKeyConfiguration{
		Key:     key,
		Type:    t,
		Maximum: &max,
	}
}

func TestAddLoginIDHandler(t *testing.T) {
	Convey("Test AddLoginIDHandler", t, func() {
		h := &AddLoginIDHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			AddLoginIDRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		authContext := authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1").
			SetVerifyInfo(map[string]bool{"user1@example.com": true}).
			MarkVerified()
		h.AuthContext = authContext
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"user-id-1": *authContext.MustAuthInfo(),
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
		h.UserVerificationProvider = userverify.NewProvider(nil, nil, &config.UserVerificationConfiguration{
			Criteria: config.UserVerificationCriteriaAll,
			LoginIDKeys: []config.UserVerificationKeyConfiguration{
				config.UserVerificationKeyConfiguration{Key: "email"},
			},
		}, nil)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("should validate request", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "id", "value": "user1"
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
							{ "kind": "General", "message": "login ID key is not allowed", "pointer": "/key" }
						]
					}
				}
			}`)
		})

		Convey("should fail if login ID is already used", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "username", "value": "user2"
			}`))
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

		Convey("should fail if there are too many login ID", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "email", "value": "user1+a@example.com"
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
								"message": "too many login IDs",
								"pointer": "",
								"details": { "key": "email", "lte": 1 }
							}
						]
					}
				}
			}`)
		})

		Convey("should add new login ID", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "username", "value": "user1"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			So(passwordAuthProvider.PrincipalMap, ShouldHaveLength, 3)
			var p password.Principal
			err := passwordAuthProvider.GetPrincipalByLoginIDWithRealm("username", "user1", password.DefaultRealm, &p)
			So(err, ShouldBeNil)
			So(p.UserID, ShouldEqual, "user-id-1")
			So(p.LoginIDKey, ShouldEqual, "username")
			So(p.LoginID, ShouldEqual, "user1")

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
						ID:   p.ID,
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

		Convey("should invalidate verify state", func() {
			passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
				[]config.LoginIDKeyConfiguration{
					newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 2),
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
				},
			)
			h.PasswordAuthProvider = passwordAuthProvider

			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "email", "value": "user1+a@example.com"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			So(authInfoStore.AuthInfoMap["user-id-1"].Verified, ShouldBeFalse)
		})

		Convey("should use empty password hash if no principal exists", func() {
			passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
				[]config.LoginIDKeyConfiguration{
					newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 2),
				},
				[]string{password.DefaultRealm},
				map[string]password.Principal{},
			)
			h.PasswordAuthProvider = passwordAuthProvider

			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"key": "email", "value": "user1+a@example.com"
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			principals, _ := passwordAuthProvider.GetPrincipalsByUserID("user-id-1")
			So(principals, ShouldHaveLength, 1)
			So(principals[0].HashedPassword, ShouldBeEmpty)
		})
	})
}
