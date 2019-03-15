package handler

import (
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/task"

	"github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestSingupHandler(t *testing.T) {
	Convey("Test SignupRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := SignupRequestPayload{
				AuthData: map[string]string{
					"username": "john.doe",
					"email":    "john.doe@example.com",
				},
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without auth data", func() {
			payload := SignupRequestPayload{
				Password: "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := SignupRequestPayload{
				AuthData: map[string]string{
					"username": "john.doe",
					"email":    "john.doe@example.com",
				},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate duplicated keys found in auth data in profile", func() {
			payload := SignupRequestPayload{
				AuthData: map[string]string{
					"username": "john.doe",
					"email":    "john.doe@example.com",
				},
				Password: "123456",
				RawProfile: map[string]interface{}{
					"username":  "john.doe",
					"firstname": "john",
					"lastname":  "doe",
				},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})

	Convey("Test SignupHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		loginIDMetadataKeys := [][]string{[]string{"email"}, []string{"username"}}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDMetadataKeys)
		anonymousAuthProvider := anonymous.NewMockProvider()
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		h := &SignupHandler{}
		h.AuthInfoStore = authInfoStore
		h.TokenStore = tokenStore
		h.PasswordChecker = passwordChecker
		h.PasswordAuthProvider = passwordAuthProvider
		h.AnonymousAuthProvider = anonymousAuthProvider
		h.AuditTrail = audit.NewMockTrail(t)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		h.TaskQueue = mockTaskQueue

		Convey("signup user with auth data", func() {
			authData := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "123456",
			}
			resp, err := h.Handle(payload)

			authResp, ok := resp.(response.AuthResponse)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)

			userID := authResp.UserID
			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth(userID, &a)
			So(a.ID, ShouldEqual, userID)
			So(a.LastLoginAt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)

			// check the token
			tokenStr := authResp.AccessToken
			token := authtoken.Token{}
			tokenStore.Get(tokenStr, &token)
			So(token.AuthInfoID, ShouldEqual, userID)
			So(!token.IsExpired(), ShouldBeTrue)

			// check user profile
			metadata := authResp.Metadata
			So(metadata["username"], ShouldEqual, "john.doe")
			So(metadata["email"], ShouldEqual, "john.doe@example.com")
		})

		Convey("auth data key combination should be unique", func() {
			authData := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "123456",
			}
			_, err := h.Handle(payload)
			So(err, ShouldBeNil)

			// change email only
			authData["email"] = "john.doe1@example.com"
			resp, err := h.Handle(payload)

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("anonymous singup is not supported yet", func() {
			payload := SignupRequestPayload{}

			resp, err := h.Handle(payload)

			authResp, ok := resp.(response.AuthResponse)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)

			userID := authResp.UserID
			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth(userID, &a)
			So(a.ID, ShouldEqual, userID)
			So(a.LastLoginAt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)

			// check the token
			tokenStr := authResp.AccessToken
			token := authtoken.Token{}
			tokenStore.Get(tokenStr, &token)
			So(token.AuthInfoID, ShouldEqual, userID)
			So(!token.IsExpired(), ShouldBeTrue)
		})

		Convey("signup with incorrect auth data", func() {
			authData := map[string]string{
				"phone": "202-111-2222",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: invalid auth data")
		})

		Convey("signup with weak password", func() {
			authData := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "1234",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "PasswordPolicyViolated: password too short")
		})

		Convey("signup with email, send welcome email", func() {
			h.WelcomeEmailEnabled = true
			authData := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "12345678",
			}
			_, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(mockTaskQueue.TasksName, ShouldResemble, []string{task.WelcomeEmailSendTaskName})

			So(mockTaskQueue.TasksParam, ShouldHaveLength, 1)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe@example.com")
			So(param.UserProfile, ShouldNotBeNil)
			So(param.UserProfile.Data["username"], ShouldEqual, "john.doe")
			So(param.UserProfile.Data["email"], ShouldEqual, "john.doe@example.com")
		})

		Convey("log audit trail when signup success", func() {
			authData := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				AuthData: authData,
				Password: "123456",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "signup")
		})
	})
}
