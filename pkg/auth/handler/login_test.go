package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/core/config"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

func TestLoginHandler(t *testing.T) {
	Convey("Test LoginRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "username",
				LoginID:    "john.doe",
				Password:   "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without login_id", func() {
			payload := LoginRequestPayload{
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeError)
		})

		Convey("validate payload without password", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "username",
				LoginID:    "john.doe",
			}
			So(payload.Validate(), ShouldBeError)
		})

		Convey("validate payload without login ID key", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe",
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate valid payload with realm", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe",
				Realm:    "admin",
				Password: "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})
	})

	Convey("Test LoginHandler", t, func() {
		realTime := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTime
		}()

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
		allowedRealms := []string{password.DefaultRealm, "admin"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeys,
			allowedRealms,
			map[string]password.Principal{
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue:    map[string]interface{}{},
				},
				"john.doe.principal.id3": password.Principal{
					ID:             "john.doe.principal.id3",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe+1@example.com",
					Realm:          "admin",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe+1@example.com",
					},
				},
			},
		)

		h := &LoginHandler{}
		h.AuthInfoStore = authInfoStore
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		identityProvider := principal.NewMockIdentityProvider(passwordAuthProvider)
		userProfileStore := userprofile.NewMockUserProfileStore()
		hookProvider := hook.NewMockProvider()
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		h.PasswordAuthProvider = passwordAuthProvider
		h.AuditTrail = coreAudit.NewMockTrail(t)
		h.HookProvider = hookProvider
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := config.MFAConfiguration{
			Enabled:     false,
			Enforcement: config.MFAEnforcementOptional,
		}
		mfaSender := mfa.NewMockSender()
		mfaProvider := mfa.NewProvider(mfaStore, mfaConfiguration, timeProvider, mfaSender)
		h.AuthnSessionProvider = authnsession.NewMockProvider(
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

		Convey("login user without login ID key", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe@example.com",
				Realm:    password.DefaultRealm,
				Password: "123456",
			}

			_, err := h.Handle(payload)
			So(err, ShouldBeNil)
		})

		Convey("login user with login_id and realm", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe+1@example.com",
				Realm:    "admin",
				Password: "123456",
			}

			_, err := h.Handle(payload)
			So(err, ShouldBeNil)
		})

		Convey("login user with incorrect realm", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe+1@example.com",
				Realm:    password.DefaultRealm,
				Password: "123456",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "invalid credentials")
		})

		Convey("login user with incorrect password", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Realm:      password.DefaultRealm,
				Password:   "wrong_password",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "invalid credentials")
		})

		Convey("login with incorrect login_id", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "phone",
				LoginID:    "202-111-2222",
				Realm:      password.DefaultRealm,
				Password:   "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "login ID key is not allowed")
		})

		Convey("login with disallowed realm", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe+1@example.com",
				Realm:    "test",
				Password: "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "realm is not allowed")
		})

		Convey("log audit trail when login success", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Realm:      password.DefaultRealm,
				Password:   "123456",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_success")
		})

		Convey("log audit trail when login fail", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Realm:      password.DefaultRealm,
				Password:   "wrong_password",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_failure")
		})
	})

	Convey("Test LoginHandler response", t, func() {
		realTime := timeNow
		now := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		timeNow = func() time.Time { return now }
		defer func() {
			timeNow = realTime
		}()

		// fixture
		userID := "john.doe.id"
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				userID: authinfo.AuthInfo{
					ID:         userID,
					Verified:   true,
					VerifyInfo: map[string]bool{},
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
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
					ClaimsValue:    map[string]interface{}{},
				},
			},
		)

		lh := &LoginHandler{}
		lh.AuthInfoStore = authInfoStore

		lh.PasswordAuthProvider = passwordAuthProvider
		identityProvider := principal.NewMockIdentityProvider(lh.PasswordAuthProvider)
		lh.AuditTrail = coreAudit.NewMockTrail(t)
		timeProvider := &coreTime.MockProvider{TimeNowUTC: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)}
		hookProvider := hook.NewMockProvider()
		lh.HookProvider = hookProvider
		profileData := map[string]map[string]interface{}{
			userID: map[string]interface{}{},
		}
		sessionProvider := session.NewMockProvider()
		sessionWriter := session.NewMockWriter()
		userProfileStore := userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		mfaStore := mfa.NewMockStore(timeProvider)
		mfaConfiguration := config.MFAConfiguration{
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

		Convey("should contains current identity", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_id_key": "email",
				"login_id": "john.doe@example.com",
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			lh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {}
					},
					"identity": {
						"id": "john.doe.principal.id1",
						"type": "password",
						"login_id_key": "email",
						"login_id": "john.doe@example.com",
						"realm": "default",
						"claims": {
							"email": "john.doe@example.com"
						}
					},
					"access_token": "access-token-%s-john.doe.principal.id1-0",
					"session_id": "%s-john.doe.principal.id1-0"
				}
			}`, userID, userID, userID))

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.SessionCreateEvent{
					Reason: coreAuth.SessionCreateReasonLogin,
					User: model.User{
						ID:         userID,
						Verified:   true,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   "john.doe.principal.id1",
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
						ID:                "john.doe.id-john.doe.principal.id1-0",
						IdentityID:        "john.doe.principal.id1",
						IdentityType:      "password",
						IdentityUpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					},
				},
			})
		})
	})
}
