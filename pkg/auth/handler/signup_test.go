package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/task"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
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
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestSingupHandler(t *testing.T) {
	Convey("Test SignupRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := SignupRequestPayload{
				LoginIDs: map[string]string{
					"username": "john.doe",
					"email":    "john.doe@example.com",
				},
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without login_id", func() {
			payload := SignupRequestPayload{
				Password: "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := SignupRequestPayload{
				LoginIDs: map[string]string{
					"username": "john.doe",
					"email":    "john.doe@example.com",
				},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload with duplicated loginIDs", func() {
			payload := SignupRequestPayload{
				LoginIDs: map[string]string{
					"username": "john.doe",
					"email":    "john.doe",
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

		loginIDsKeyWhitelist := []string{"email", "username"}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeyWhitelist)
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

		Convey("signup user with login_id", func() {
			loginIDs := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				LoginIDs: loginIDs,
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
			So(authResp.LoginIDs["username"], ShouldEqual, "john.doe")
			So(authResp.LoginIDs["email"], ShouldEqual, "john.doe@example.com")
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

		Convey("signup with incorrect login_id", func() {
			loginIDs := map[string]string{
				"phone": "202-111-2222",
			}
			payload := SignupRequestPayload{
				LoginIDs: loginIDs,
				Password: "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: invalid login_ids")
		})

		Convey("signup with weak password", func() {
			loginIDs := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				LoginIDs: loginIDs,
				Password: "1234",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "PasswordPolicyViolated: password too short")
		})

		Convey("signup with email, send welcome email", func() {
			h.WelcomeEmailEnabled = true
			loginIDs := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				LoginIDs: loginIDs,
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
			loginIDs := map[string]string{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			}
			payload := SignupRequestPayload{
				LoginIDs: loginIDs,
				Password: "123456",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "signup")
		})
	})

	Convey("Test SignupHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		loginIDsKeyWhitelist := []string{"email", "username"}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeyWhitelist)
		anonymousAuthProvider := anonymous.NewMockProvider()
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		sh := &SignupHandler{}
		sh.AuthInfoStore = authInfoStore
		sh.TokenStore = tokenStore
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.AnonymousAuthProvider = anonymousAuthProvider
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("duplicated user error format", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"username": "john.doe"
				},
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"username": "john.doe"
				},
				"password": "1234567"
			}`))
			resp = httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 409)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Duplicated",
					"code": 109,
					"message": "user duplicated"
				}
			}
			`)
		})
	})

	Convey("Test SignupHandler response", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		loginIDsKeyWhitelist := []string{}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeyWhitelist)
		anonymousAuthProvider := anonymous.NewMockProvider()

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		sh := &SignupHandler{}
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.AnonymousAuthProvider = anonymousAuthProvider
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should contains multiple loginIDs", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"email": "john.doe@example.com",
					"username": "john.doe"
				},
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			var p password.Principal
			err := sh.PasswordAuthProvider.GetPrincipalByLoginID("email", "john.doe@example.com", &p)
			So(err, ShouldBeNil)
			userID := p.UserID
			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "%s",
					"verified": false,
					"verify_info": {},
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s",
					"login_ids": {
						"email":"john.doe@example.com",
						"username":"john.doe"
					},
					"metadata": {}
				}
			}`,
				userID,
				token.AccessToken,
				userID,
				userID))
		})
	})
}
