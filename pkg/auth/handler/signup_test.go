package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignupHandler(t *testing.T) {
	Convey("Test SignupHandler", t, func() {
		realTime := timeNow
		now := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		timeNow = func() time.Time { return now }
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
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SignupRequestSchema,
		)
		sh.Validator = validator
		sh.AuthInfoStore = authInfoStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStore()
		identityProvider := principal.NewMockIdentityProvider(passwordAuthProvider)
		mockOAuthProvider := oauth.NewMockProvider([]*oauth.Principal{
			&oauth.Principal{
				ID:           "john.doe.id",
				UserID:       "john.doe.id",
				ProviderType: "google",
				ProviderKeys: map[string]interface{}{},
				ClaimsValue: map[string]interface{}{
					"email": "john.doe@example.com",
				},
			},
		})
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.IdentityProvider = identityProvider
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userProfileStore
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		sh.WelcomeEmailEnabled = true
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}

		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		sh.AuthnSessionProvider = authnsession.NewMockProvider(
			mfaConfiguration,
			timeProvider,
			mfaProvider,
			authInfoStore,
			sessionProvider,
			sessionWriter,
			identityProvider,
			hookProvider,
			userProfileStore,
		)
		Convey("should reject request without login ID", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "EntryAmount",
								"pointer": "/login_ids",
								"message": "Array must have at least 1 items",
								"details": { "gte": 1 }
							}
						]
					}
				}
			}`)
		})
		Convey("should reject request with duplicated login ID", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" },
					{ "key": "email", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "General",
								"pointer": "/login_ids/1/value",
								"message": "duplicated login ID"
							}
						]
					}
				}
			}`)
		})

		Convey("abort if user duplicate with oauth", func() {
			sh.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider, mockOAuthProvider)
			sh.AuthnSessionProvider = authnsession.NewMockProvider(
				mfaConfiguration,
				timeProvider,
				mfaProvider,
				authInfoStore,
				sessionProvider,
				sessionWriter,
				identityProvider,
				hookProvider,
				userProfileStore,
			)
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe@example.com" },
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 409)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is used by another user",
					"code": 409
				}
			}`)
		})

		Convey("singup with on_user_duplicate == create", func() {
			sh.AuthConfiguration = config.AuthConfiguration{
				OnUserDuplicateAllowCreate: true,
			}
			sh.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider, mockOAuthProvider)
			sh.AuthnSessionProvider = authnsession.NewMockProvider(
				mfaConfiguration,
				timeProvider,
				mfaProvider,
				authInfoStore,
				sessionProvider,
				sessionWriter,
				identityProvider,
				hookProvider,
				userProfileStore,
			)
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe@example.com" },
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456",
				"on_user_duplicate": "create"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
		})

		Convey("signup user with login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "john.doe@example.com" },
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			var p password.Principal
			err := sh.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("email", "john.doe@example.com", password.DefaultRealm, &p)
			So(err, ShouldBeNil)
			var p2 password.Principal
			err = sh.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("username", "john.doe", password.DefaultRealm, &p2)
			So(err, ShouldBeNil)

			userID := p.UserID
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
					"access_token": "access-token-%s-%s-0",
					"session_id": "%s-%s-0"
				}
			}`, userID, p.ID, userID, p.ID, userID, p.ID))

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserCreateEvent{
					User: model.User{
						ID:          userID,
						LastLoginAt: &now,
						Verified:    false,
						Disabled:    false,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identities: []model.Identity{
						model.Identity{
							ID:   p.ID,
							Type: "password",
							Attributes: principal.Attributes{
								"login_id_key": "email",
								"login_id":     "john.doe@example.com",
								"realm":        "default",
							},
							Claims: principal.Claims{
								"email": "john.doe@example.com",
							},
						},
						model.Identity{
							ID:   p2.ID,
							Type: "password",
							Attributes: principal.Attributes{
								"login_id_key": "username",
								"login_id":     "john.doe",
								"realm":        "default",
							},
							Claims: principal.Claims{},
						},
					},
				},
				event.SessionCreateEvent{
					Reason: coreAuth.SessionCreateReasonSignup,
					User: model.User{
						ID:          userID,
						LastLoginAt: &now,
						Verified:    false,
						Disabled:    false,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   p.ID,
						Type: "password",
						Attributes: principal.Attributes{
							"login_id_key": "email",
							"login_id":     "john.doe@example.com",
							"realm":        "default",
						},
						Claims: principal.Claims{
							"email": "john.doe@example.com",
						},
					},
					Session: model.Session{
						ID:                fmt.Sprintf("%s-%s-0", userID, p.ID),
						IdentityID:        p.ID,
						IdentityType:      "password",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			var p password.Principal
			err := sh.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("email", "john.doe@example.com", "admin", &p)
			So(err, ShouldBeNil)
			userID := p.UserID
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
					"access_token": "access-token-%s-%s-0",
					"session_id": "%s-%s-0"
				}
			}`, userID, p.ID, userID, p.ID, userID, p.ID))

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserCreateEvent{
					User: model.User{
						ID:          userID,
						LastLoginAt: &now,
						Verified:    false,
						Disabled:    false,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identities: []model.Identity{
						model.Identity{
							ID:   p.ID,
							Type: "password",
							Attributes: principal.Attributes{
								"login_id_key": "email",
								"login_id":     "john.doe@example.com",
								"realm":        "admin",
							},
							Claims: principal.Claims{
								"email": "john.doe@example.com",
							},
						},
					},
				},
				event.SessionCreateEvent{
					Reason: coreAuth.SessionCreateReasonSignup,
					User: model.User{
						ID:          userID,
						LastLoginAt: &now,
						Verified:    false,
						Disabled:    false,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   p.ID,
						Type: "password",
						Attributes: principal.Attributes{
							"login_id_key": "email",
							"login_id":     "john.doe@example.com",
							"realm":        "admin",
						},
						Claims: principal.Claims{
							"email": "john.doe@example.com",
						},
					},
					Session: model.Session{
						ID:                fmt.Sprintf("%s-%s-0", userID, p.ID),
						IdentityID:        p.ID,
						IdentityType:      "password",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
		})

		Convey("signup with incorrect login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "phone", "value": "202-111-2222" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "General",
								"pointer": "/login_ids",
								"message": "login ID key is not allowed"
							}
						]
					}
				}
			}
			`)
		})

		Convey("signup with invalid login_id", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "email", "value": "202-111-2222" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "StringFormat",
								"pointer": "/login_ids/0/value",
								"message": "invalid login ID format",
								"details": { "format": "email" }
							}
						]
					}
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Invalid",
					"reason": "PasswordPolicyViolated",
					"message": "password policy violated",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "PasswordTooShort", "min_length": 6, "pw_length": 4 }
						]
					}
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
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
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
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

		passwordChecker := &authAudit.PasswordChecker{
			PwMinLength: 6,
		}

		sh := &SignupHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SignupRequestSchema,
		)
		sh.Validator = validator
		sh.AuthInfoStore = authInfoStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStore()
		identityProvider := principal.NewMockIdentityProvider(passwordAuthProvider)
		sh.PasswordChecker = passwordChecker
		sh.PasswordAuthProvider = passwordAuthProvider
		sh.IdentityProvider = identityProvider
		sh.AuditTrail = audit.NewMockTrail(t)
		sh.UserProfileStore = userProfileStore
		sh.Logger = logrus.NewEntry(logrus.New())
		mockTaskQueue := async.NewMockQueue()
		sh.TaskQueue = mockTaskQueue
		sh.TxContext = db.NewMockTxContext()
		hookProvider := hook.NewMockProvider()
		sh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}

		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		sh.AuthnSessionProvider = authnsession.NewMockProvider(
			mfaConfiguration,
			timeProvider,
			mfaProvider,
			authInfoStore,
			sessionProvider,
			sessionWriter,
			identityProvider,
			hookProvider,
			userProfileStore,
		)

		Convey("duplicated user error format", func(c C) {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)

			req, _ = http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [
					{ "key": "username", "value": "john.doe" }
				],
				"password": "1234567"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp = httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 409)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is used by another user",
					"code": 409
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
			req.Header.Set("Content-Type", "application/json")
			resp = httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 409)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "AlreadyExists",
					"reason": "LoginIDAlreadyUsed",
					"message": "login ID is used by another user",
					"code": 409
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
			req.Header.Set("Content-Type", "application/json")
			resp = httptest.NewRecorder()
			sh.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `
			{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "General",
								"pointer": "/realm",
								"message": "realm is not a valid realm"
							}
						]
					}
				}
			}
			`)
		})
	})
}
