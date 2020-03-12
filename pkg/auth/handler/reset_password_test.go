package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResetPasswordHandler(t *testing.T) {
	Convey("Test ResetPasswordHandler", t, func() {
		// fixture
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		one := 1
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{Key: "email", Maximum: &one},
			config.LoginIDKeyConfiguration{Key: "username", Maximum: &one},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:             "john.doe.principal.id0",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}
		mockTaskQueue := async.NewMockQueue()

		h := &ResetPasswordHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			ResetPasswordRequestSchema,
		)
		h.TxContext = db.NewMockTxContext()
		h.Validator = validator
		h.AuthInfoStore = authInfoStore
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.PasswordChecker = passwordChecker
		h.PasswordAuthProvider = passwordAuthProvider
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider
		h.TaskQueue = mockTaskQueue

		Convey("should reject invalid request", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "Required", "message": "password is required", "pointer": "/password" },
							{ "kind": "Required", "message": "user_id is required", "pointer": "/user_id" }
						]
					}
				}
			}`)
		})

		Convey("should reset password by user id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"password": "234567"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			// should update all principals of a user
			principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID("john.doe.id")
			So(err, ShouldBeNil)
			for _, p := range principals {
				So(p.IsSamePassword("234567"), ShouldEqual, true)
			}

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, spec.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, spec.PwHousekeeperTaskParam{
				AuthID: "john.doe.id",
			})

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.PasswordUpdateEvent{
					Reason: event.PasswordUpdateReasonAdministrative,
					User: model.User{
						ID:         "john.doe.id",
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
				},
			})
		})

		Convey("should not reset password by wrong user id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id.wrong",
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "UserNotFound",
					"message": "user not found",
					"code": 404
				}
			}`)
		})

		Convey("should not reset password with password violates password policy", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"user_id": "john.doe.id",
				"password": "1234"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "PasswordPolicyViolated",
					"message": "password policy violated",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "PasswordTooShort", "min_length" : 6, "pw_length": 4 }
						]
					}
				}
			}`)
		})
	})
}
