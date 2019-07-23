package handler

import (
	"testing"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResetPasswordPayload(t *testing.T) {
	Convey("Test ResetPasswordRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := ResetPasswordRequestPayload{
				AuthInfoID: "1",
				Password:   "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without auth_id", func() {
			payload := ResetPasswordRequestPayload{
				Password: "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := ResetPasswordRequestPayload{
				AuthInfoID: "1",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}
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
		zero := 0
		one := 1
		loginIDsKeys := map[string]config.LoginIDKeyConfiguration{
			"email":    config.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
			"username": config.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
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
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)
		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}
		mockTaskQueue := async.NewMockQueue()

		h := &ResetPasswordHandler{}
		h.AuthInfoStore = authInfoStore
		h.TokenStore = tokenStore
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.PasswordChecker = passwordChecker
		h.PasswordAuthProvider = passwordAuthProvider
		h.AuditTrail = audit.NewMockTrail(t)
		h.HookProvider = hook.NewMockProvider()
		h.TaskQueue = mockTaskQueue

		Convey("should reset password by user id", func() {
			userID := "john.doe.id"
			newPassword := "234567"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   newPassword,
			}

			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})

			// should update all principals of a user
			principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(userID)
			So(err, ShouldBeNil)
			for _, p := range principals {
				So(p.IsSamePassword(newPassword), ShouldEqual, true)
			}

			// should enqueue pw housekeeper task
			So(mockTaskQueue.TasksName[0], ShouldEqual, task.PwHousekeeperTaskName)
			So(mockTaskQueue.TasksParam[0], ShouldResemble, task.PwHousekeeperTaskParam{
				AuthID: userID,
			})
		})

		Convey("should not reset password by wrong user id", func() {
			userID := "john.doe.id.wrong"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "123456",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "ResourceNotFound: User not found")
		})

		Convey("should not reset password with password violates password policy", func() {
			userID := "john.doe.id"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "1234",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "PasswordPolicyViolated: password too short")
		})

		Convey("should have audit trail when reset password", func() {
			userID := "john.doe.id"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "123456",
			}

			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "reset_password")
		})
	})
}
