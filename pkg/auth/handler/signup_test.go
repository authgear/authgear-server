package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/response"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"

	"github.com/skygeario/skygear-server/pkg/auth"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/task"
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
		sh.WelcomeEmailEnabled = true
		hookExecutor := hook.NewMockExecutorImpl(map[string]hook.MockExecutorResult{})
		authHooks := []config.AuthHook{}
		sh.AuthHooksStore = hook.NewHookProvider(authHooks, hookExecutor, logrus.NewEntry(logrus.New()))
		h := auth.HookHandlerToAPIHandler(sh, sh.TxContext)

		Convey("signup user with login_id", func() {
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

		Convey("support anonymous singup", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader("{}"))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			userID := anonymousAuthProvider.Principals[0].UserID
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
					"metadata": {}
				}
			}`,
				userID,
				token.AccessToken,
				userID,
				userID))
		})

		Convey("signup with incorrect login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"phone": "202-111-2222"
				},
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "InvalidArgument",
					"code": 108,
					"info":{
						"arguments":["login_ids"]
					},
					"message": "invalid login_ids","name":"InvalidArgument"
				}
			}
			`)
		})

		Convey("signup with weak password", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"username": "john.doe",
					"email":    "john.doe@example.com"
				},
				"password": "1234"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "PasswordPolicyViolated",
					"code": 126,
					"info":{
							"min_length": 6,
							"pw_length": 4,
							"reason": "PasswordTooShort"
					},
					"message": "password too short"
				}
			}
			`)
		})

		Convey("signup with email, send welcome email", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"username": "john.doe",
					"email":    "john.doe@example.com"
				},
				"password": "12345678"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			So(mockTaskQueue.TasksName, ShouldResemble, []string{task.WelcomeEmailSendTaskName})
			So(mockTaskQueue.TasksParam, ShouldHaveLength, 1)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe@example.com")
			So(param.User, ShouldNotBeNil)
			So(param.User.LoginIDs["username"], ShouldEqual, "john.doe")
			So(param.User.LoginIDs["email"], ShouldEqual, "john.doe@example.com")
		})

		Convey("log audit trail when signup success", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": {
					"username": "john.doe",
					"email":    "john.doe@example.com"
				},
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			mockTrail, _ := sh.AuditTrail.(*audit.MockTrail)
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
		hookExecutor := hook.NewMockExecutorImpl(map[string]hook.MockExecutorResult{})
		authHooks := []config.AuthHook{}
		sh.AuthHooksStore = hook.NewHookProvider(authHooks, hookExecutor, logrus.NewEntry(logrus.New()))
		h := auth.HookHandlerToAPIHandler(sh, sh.TxContext)

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

	Convey("Test signup hooks", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

		loginIDsKeyWhitelist := []string{"email", "username"}
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
		sh.WelcomeEmailEnabled = true

		Convey("should invoke before signup hook", func(c C) {
			hookExecutor := hook.NewMockExecutorImpl(map[string]hook.MockExecutorResult{
				"before_signup_hook_url": hook.MockExecutorResult{
					User: response.User{
						Metadata: userprofile.Data{
							"name": "john.doe",
						},
					},
					Error: nil,
				},
			})
			authHooks := []config.AuthHook{
				config.AuthHook{
					Event: hook.BeforeSignup,
					URL:   "before_signup_hook_url",
				},
			}
			sh.AuthHooksStore = hook.NewHookProvider(authHooks, hookExecutor, logrus.NewEntry(logrus.New()))
			h := auth.HookHandlerToAPIHandler(sh, sh.TxContext)

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
					"metadata": {
						"name": "john.doe"
					}
				}
			}`,
				userID,
				token.AccessToken,
				userID,
				userID))
		})

		Convey("should stop signup if hook throws error", func(c C) {
			hookExecutor := hook.NewMockExecutorImpl(map[string]hook.MockExecutorResult{
				"after_signup_hook_url": hook.MockExecutorResult{
					Error: errors.New("after_signup_fail"),
				},
			})
			authHooks := []config.AuthHook{
				config.AuthHook{
					Event: hook.AfterSignup,
					URL:   "after_signup_hook_url",
				},
			}
			sh.AuthHooksStore = hook.NewHookProvider(authHooks, hookExecutor, logrus.NewEntry(logrus.New()))
			h := auth.HookHandlerToAPIHandler(sh, sh.TxContext)

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

			So(resp.Code, ShouldEqual, 500)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "UnexpectedError",
					"code": 10000,
					"message": "after_signup_fail"
				}
			}
			`)
		})
	})
}
