package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestSingupHandler(t *testing.T) {
	Convey("Test SignupRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := SignupRequestPayload{
				LoginIDs: []password.LoginID{
					password.LoginID{Key: "username", Value: "john.doe"},
					password.LoginID{Key: "email", Value: "john.doe@example.com"},
				},
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate valid payload with realm", func() {
			payload := SignupRequestPayload{
				LoginIDs: []password.LoginID{
					password.LoginID{Key: "username", Value: "john.doe"},
					password.LoginID{Key: "email", Value: "john.doe@example.com"},
				},
				Realm:    "admin",
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
				LoginIDs: []password.LoginID{
					password.LoginID{Key: "username", Value: "john.doe"},
					password.LoginID{Key: "email", Value: "john.doe@example.com"},
				},
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload with duplicated loginIDs", func() {
			payload := SignupRequestPayload{
				LoginIDs: []password.LoginID{
					password.LoginID{Key: "username", Value: "john.doe"},
					password.LoginID{Key: "email", Value: "john.doe"},
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

		zero := 0
		two := 2
		loginIDsKeys := map[string]config.LoginIDKeyConfiguration{
			"email": config.LoginIDKeyConfiguration{
				Type:    config.LoginIDKeyType(metadata.Email),
				Minimum: &zero,
				Maximum: &two,
			},
			"username": config.LoginIDKeyConfiguration{
				Type:    config.LoginIDKeyTypeRaw,
				Minimum: &zero,
				Maximum: &two,
			},
		}
		allowedRealms := []string{password.DefaultRealm, "admin"}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeys, allowedRealms)

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		sh := &SignupHandler{}
		sh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		sh.TokenStore = mockTokenStore
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.PasswordAuthProvider)
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		sh.WelcomeEmailEnabled = true
		executor := hook.ExecutorImpl{}
		hooks := []config.Hook{}
		sh.HookStore = hook.NewHookProvider(hooks, executor, logrus.NewEntry(logrus.New()), "")
		h := auth.HookHandlerToHandler(sh, sh.TxContext)

		Convey("signup user with login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe@example.com" },
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			var p password.Principal
			err := sh.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("email", "john.doe@example.com", password.DefaultRealm, &p)
			So(err, ShouldBeNil)
			userID := p.UserID
			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "%s",
						"type": "password",
						"login_id_key": "email",
						"login_id": "john.doe@example.com",
						"realm": "default",
						"claims": {
							"email": "john.doe@example.com"
						}
					},
					"access_token": "%s"
				}
			}`,
				userID,
				p.ID,
				token.AccessToken))
		})

		Convey("signup user with login_id with realm", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe@example.com" }
				],
				"realm": "admin",
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			var p password.Principal
			err := sh.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("email", "john.doe@example.com", "admin", &p)
			So(err, ShouldBeNil)
			userID := p.UserID
			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "%s",
						"type": "password",
						"login_id_key": "email",
						"login_id": "john.doe@example.com",
						"realm": "admin",
						"claims": {
							"email": "john.doe@example.com"
						}
					},
					"access_token": "%s"
				}
			}`,
				userID,
				p.ID,
				token.AccessToken))
		})

		Convey("signup with incorrect login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "phone", "value": "202-111-2222" }
				],
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
						"arguments":["phone"]
					},
					"message": "login ID key is not allowed","name":"InvalidArgument"
				}
			}
			`)
		})

		Convey("signup with weak password", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" },
					{ "key": "email", "value": "john.doe@example.com" }
				],
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

		Convey("signup with email, send welcome email to first login ID", func() {
			sh.WelcomeEmailDestination = config.WelcomeEmailDestinationFirst
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe+1@example.com" },
					{ "key": "username", "value": "john.doe" },
					{ "key": "email", "value": "john.doe+2@example.com" }
				],
				"password": "12345678"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			So(mockTaskQueue.TasksName, ShouldResemble, []string{task.WelcomeEmailSendTaskName})
			So(mockTaskQueue.TasksParam, ShouldHaveLength, 1)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe+1@example.com")
			So(param.User, ShouldNotBeNil)
		})

		Convey("signup with email, send welcome email to all login IDs", func() {
			sh.WelcomeEmailDestination = config.WelcomeEmailDestinationAll
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe+1@example.com" },
					{ "key": "username", "value": "john.doe" },
					{ "key": "email", "value": "john.doe+2@example.com" }
				],
				"password": "12345678"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			So(mockTaskQueue.TasksName, ShouldResemble, []string{task.WelcomeEmailSendTaskName, task.WelcomeEmailSendTaskName})
			So(mockTaskQueue.TasksParam, ShouldHaveLength, 2)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe+1@example.com")
			So(param.User, ShouldNotBeNil)
			param, _ = mockTaskQueue.TasksParam[1].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "john.doe+2@example.com")
			So(param.User, ShouldNotBeNil)
		})

		Convey("log audit trail when signup success", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" },
					{ "key": "email", "value": "john.doe@example.com" }
				],
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

		zero := 0
		one := 1
		loginIDsKeys := map[string]config.LoginIDKeyConfiguration{
			"email":    config.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
			"username": config.LoginIDKeyConfiguration{Minimum: &zero, Maximum: &one},
		}
		allowedRealms := []string{password.DefaultRealm, "admin"}
		authInfoStore := authinfo.NewMockStore()
		passwordAuthProvider := password.NewMockProvider(loginIDsKeys, allowedRealms)
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		sh := &SignupHandler{}
		sh.AuthInfoStore = authInfoStore
		sh.TokenStore = tokenStore
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.IdentityProvider = principal.NewMockIdentityProvider(sh.PasswordAuthProvider)
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userprofile.NewMockUserProfileStore()
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		executor := hook.ExecutorImpl{}
		hooks := []config.Hook{}
		sh.HookStore = hook.NewHookProvider(hooks, executor, logrus.NewEntry(logrus.New()), "")
		h := auth.HookHandlerToHandler(sh, sh.TxContext)

		Convey("duplicated user error format", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
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

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
				"realm": "admin",
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

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
				"realm": "test",
				"password": "1234567"
			}`))
			resp = httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "InvalidArgument",
					"code": 108,
					"info":{
						"arguments":["realm"]
					},
					"message": "realm is not allowed"
				}
			}
			`)
		})
	})
}
