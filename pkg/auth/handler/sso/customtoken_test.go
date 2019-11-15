package sso

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/customtoken"
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
)

func TestCustomTokenLoginHandler(t *testing.T) {
	realTime := timeNow
	now := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	timeNow = func() time.Time { return now }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test CustomTokenLoginHandler", t, func() {
		lh := &CustomTokenLoginHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			CustomTokenLoginRequestSchema,
		)
		lh.Validator = validator
		issuer := "myissuer"
		audience := "myaudience"
		mockPasswordProvider := password.NewMockProvider(
			nil,
			[]string{password.DefaultRealm},
		)
		lh.PasswordAuthProvider = mockPasswordProvider
		lh.CustomTokenConfiguration = config.CustomTokenConfiguration{
			Enabled:  true,
			Issuer:   issuer,
			Audience: audience,
		}
		lh.TxContext = db.NewMockTxContext()
		customTokenAuthProvider := customtoken.NewMockProviderWithPrincipalMap("ssosecret", map[string]customtoken.Principal{
			"uuid-chima-token": customtoken.Principal{
				ID:               "uuid-chima-token",
				TokenPrincipalID: "chima.customtoken.id",
				UserID:           "chima",
			},
		})
		lh.CustomTokenAuthProvider = customTokenAuthProvider
		identityProvider := principal.NewMockIdentityProvider(customTokenAuthProvider)
		lh.IdentityProvider = identityProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"chima": authinfo.AuthInfo{
					ID: "chima",
				},
			},
		)
		lh.AuthInfoStore = authInfoStore
		userProfileStore := userprofile.NewMockUserProfileStore()
		userProfileStore.Data = map[string]map[string]interface{}{}
		userProfileStore.Data["chima"] = map[string]interface{}{
			"name":  "chima",
			"email": "chima@skygear.io",
		}
		userProfileStore.TimeNowfunc = timeNow
		lh.UserProfileStore = userProfileStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		lh.AuditTrail = audit.NewMockTrail(t)
		lh.WelcomeEmailEnabled = true
		mockTaskQueue := async.NewMockQueue()
		lh.TaskQueue = mockTaskQueue
		hookProvider := hook.NewMockProvider()
		lh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := &config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		lh.AuthnSessionProvider = authnsession.NewMockProvider(
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

		iat := time.Now().UTC()
		exp := iat.Add(time.Hour * 1)

		Convey("create user account with custom token", func(c C) {
			claims := customtoken.SSOCustomTokenClaims{
				"iss":   issuer,
				"aud":   audience,
				"iat":   float64(iat.Unix()),
				"exp":   float64(exp.Unix()),
				"sub":   "otherid1",
				"email": "John@skygear.io",
			}
			tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("otherid1")

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": false,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "2006-01-02T15:04:05Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "%s",
						"type": "custom_token",
						"provider_user_id": "otherid1",
						"raw_profile": {
							"aud": "myaudience",
							"email": "John@skygear.io",
							"iat": %d,
							"exp": %d,
							"iss": "myissuer",
							"sub": "otherid1"
						},
						"claims": {
							"email": "John@skygear.io"
						}
					},
					"access_token": "access-token-%s-%s-0",
					"session_id": "%s-%s-0"
				}
			}`,
				p.UserID,
				p.ID,
				iat.Unix(),
				exp.Unix(),
				p.UserID,
				p.ID,
				p.UserID,
				p.ID))

			mockTrail, _ := lh.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "signup")

			So(mockTaskQueue.TasksParam, ShouldHaveLength, 1)
			param, _ := mockTaskQueue.TasksParam[0].(task.WelcomeEmailSendTaskParam)
			So(param.Email, ShouldEqual, "John@skygear.io")
			So(param.User, ShouldNotBeNil)
			So(param.User.Metadata, ShouldResemble, userprofile.Data{})

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.UserCreateEvent{
					User: model.User{
						ID:          p.UserID,
						CreatedAt:   now,
						LastLoginAt: &now,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identities: []model.Identity{
						model.Identity{
							ID:   p.ID,
							Type: "custom_token",
							Attributes: principal.Attributes{
								"provider_user_id": "otherid1",
								"raw_profile":      claims,
							},
							Claims: principal.Claims{
								"email": "John@skygear.io",
							},
						},
					},
				},
				event.SessionCreateEvent{
					Reason: coreAuth.SessionCreateReasonSignup,
					User: model.User{
						ID:          p.UserID,
						CreatedAt:   now,
						LastLoginAt: &now,
						VerifyInfo:  map[string]bool{},
						Metadata:    userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   p.ID,
						Type: "custom_token",
						Attributes: principal.Attributes{
							"provider_user_id": "otherid1",
							"raw_profile":      claims,
						},
						Claims: principal.Claims{
							"email": "John@skygear.io",
						},
					},
					Session: model.Session{
						ID:                fmt.Sprintf("%s-%s-0", p.UserID, p.ID),
						IdentityID:        p.ID,
						IdentityType:      "custom_token",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
		})

		Convey("does not update user account with custom token", func(c C) {
			claims := customtoken.SSOCustomTokenClaims{
				"iss":   issuer,
				"aud":   audience,
				"iat":   float64(time.Now().Unix()),
				"exp":   float64(time.Now().Add(time.Hour * 1).Unix()),
				"sub":   "chima.customtoken.id",
				"email": "John@skygear.io",
			}
			tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("chima.customtoken.id")
			So(p.UserID, ShouldEqual, "chima")

			profile, _ := lh.UserProfileStore.GetUserProfile(p.UserID)
			So(profile.Data, ShouldResemble, userprofile.Data{
				"name":  "chima",
				"email": "chima@skygear.io",
			})

			So(mockTaskQueue.TasksParam, ShouldHaveLength, 0)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.SessionCreateEvent{
					Reason: coreAuth.SessionCreateReasonLogin,
					User: model.User{
						ID:         "chima",
						CreatedAt:  now,
						VerifyInfo: map[string]bool{},
						Metadata: userprofile.Data{
							"name":  "chima",
							"email": "chima@skygear.io",
						},
					},
					Identity: model.Identity{
						ID:   "uuid-chima-token",
						Type: "custom_token",
						Attributes: principal.Attributes{
							"provider_user_id": "chima.customtoken.id",
							"raw_profile":      claims,
						},
						Claims: principal.Claims{
							"email": "John@skygear.io",
						},
					},
					Session: model.Session{
						ID:                "chima-uuid-chima-token-0",
						IdentityID:        "uuid-chima-token",
						IdentityType:      "custom_token",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
		})

		Convey("check whether token is invalid", func(c C) {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					"iss": issuer,
					"aud": audience,
					"iat": time.Now().Add(-time.Hour * 1).Unix(),
					"exp": time.Now().Add(-time.Minute * 30).Unix(),
					"sub": "otherid1",
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 401)

			mockTrail, _ := lh.AuditTrail.(*audit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_failure")
		})

		Convey("should return error if disabled", func() {
			tokenString, err := jwt.NewWithClaims(
				jwt.SigningMethodHS256,
				customtoken.SSOCustomTokenClaims{
					"iss":   issuer,
					"aud":   audience,
					"iat":   time.Now().Unix(),
					"exp":   time.Now().Add(time.Hour * 1).Unix(),
					"sub":   "otherid1",
					"email": "John@skygear.io",
				},
			).SignedString([]byte("ssosecret"))
			So(err, ShouldBeNil)

			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			lhh := lh
			lhh.CustomTokenConfiguration.Enabled = false

			lhh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 404)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "NotFound",
					"reason": "NotFound",
					"message": "custom token is disabled",
					"code": 404
				}
			}`)
		})
	})

	Convey("Test OnUserDuplicate", t, func() {
		lh := &CustomTokenLoginHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			CustomTokenLoginRequestSchema,
		)
		lh.Validator = validator
		issuer := "myissuer"
		audience := "myaudience"
		zero := 0
		one := 1
		loginIDsKeys := []config.LoginIDKeyConfiguration{
			config.LoginIDKeyConfiguration{
				Key:     "email",
				Type:    config.LoginIDKeyType(metadata.Email),
				Minimum: &zero,
				Maximum: &one,
			},
		}
		allowedRealms := []string{password.DefaultRealm}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:             "john.doe.principal.id",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          "default",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
			},
		)
		lh.PasswordAuthProvider = passwordAuthProvider
		lh.CustomTokenConfiguration = config.CustomTokenConfiguration{
			Enabled:                    true,
			Issuer:                     issuer,
			Audience:                   audience,
			OnUserDuplicateAllowMerge:  true,
			OnUserDuplicateAllowCreate: true,
		}
		lh.TxContext = db.NewMockTxContext()
		customTokenAuthProvider := customtoken.NewMockProviderWithPrincipalMap("ssosecret", map[string]customtoken.Principal{})
		lh.CustomTokenAuthProvider = customTokenAuthProvider
		identityProvider := principal.NewMockIdentityProvider(lh.CustomTokenAuthProvider, passwordAuthProvider)
		lh.IdentityProvider = identityProvider
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID:         "john.doe.id",
					VerifyInfo: map[string]bool{},
				},
			},
		)
		lh.AuthInfoStore = authInfoStore
		userProfileStore := userprofile.NewMockUserProfileStoreByData(map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{},
		})
		userProfileStore.TimeNowfunc = timeNow
		lh.UserProfileStore = userProfileStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		lh.AuditTrail = audit.NewMockTrail(t)
		lh.WelcomeEmailEnabled = true
		mockTaskQueue := async.NewMockQueue()
		lh.TaskQueue = mockTaskQueue
		hookProvider := hook.NewMockProvider()
		lh.HookProvider = hookProvider
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := &config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		lh.AuthnSessionProvider = authnsession.NewMockProvider(
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

		iat := time.Now().UTC()
		exp := iat.Add(time.Hour * 1)

		tokenString, err := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			customtoken.SSOCustomTokenClaims{
				"iss":   issuer,
				"aud":   audience,
				"iat":   iat.Unix(),
				"exp":   exp.Unix(),
				"sub":   "otherid1",
				"email": "john.doe@example.com",
			},
		).SignedString([]byte("ssosecret"))
		So(err, ShouldBeNil)

		Convey("OnUserDuplicate == abort", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

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
		})

		Convey("OnUserDuplicate == merge", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s",
				"on_user_duplicate": "merge"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("otherid1")

			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`
			{
				"result": {
					"user": {
						"created_at": "2006-01-02T15:04:05Z",
						"id": "john.doe.id",
						"is_disabled": false,
						"is_verified": false,
						"metadata": {},
						"verify_info": {}
					},
					"identity": {
						"claims": {
							"email": "john.doe@example.com"
						},
						"id": "%s",
						"provider_user_id": "otherid1",
						"raw_profile": {
							"iss": "myissuer",
							"aud": "myaudience",
							"sub": "otherid1",
							"iat": %d,
							"exp": %d,
							"email": "john.doe@example.com"
						},
						"type": "custom_token"
					},
					"access_token": "access-token-john.doe.id-%s-0",
					"session_id": "john.doe.id-%s-0"
				}
			}
			`, p.ID,
				iat.Unix(),
				exp.Unix(),
				p.ID,
				p.ID))
		})

		Convey("OnUserDuplicate == create", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(fmt.Sprintf(`
			{
				"token": "%s",
				"on_user_duplicate": "create"
			}`, tokenString)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)

			p, _ := lh.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID("otherid1")

			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`
			{
				"result": {
					"user": {
						"created_at": "2006-01-02T15:04:05Z",
						"last_login_at": "2006-01-02T15:04:05Z",
						"id": "%s",
						"is_disabled": false,
						"is_verified": false,
						"metadata": {},
						"verify_info": {}
					},
					"identity": {
						"claims": {
							"email": "john.doe@example.com"
						},
						"id": "%s",
						"provider_user_id": "otherid1",
						"raw_profile": {
							"iss": "myissuer",
							"aud": "myaudience",
							"sub": "otherid1",
							"iat": %d,
							"exp": %d,
							"email": "john.doe@example.com"
						},
						"type": "custom_token"
					},
					"access_token": "access-token-%s-%s-0",
					"session_id": "%s-%s-0"
				}
			}
			`, p.UserID, p.ID, iat.Unix(), exp.Unix(), p.UserID, p.ID, p.UserID, p.ID))
		})
	})
}
