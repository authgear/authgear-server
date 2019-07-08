package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/config"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "username",
				LoginID:    "john.doe",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
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
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id3": password.Principal{
					ID:             "john.doe.principal.id3",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe+1@example.com",
					Realm:          "admin",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)

		h := &LoginHandler{}
		h.AuthInfoStore = authInfoStore
		h.TokenStore = tokenStore
		h.PasswordAuthProvider = passwordAuthProvider
		h.IdentityProvider = principal.NewMockIdentityProvider(h.PasswordAuthProvider)
		h.AuditTrail = coreAudit.NewMockTrail(t)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()

		Convey("login user with login_id", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Realm:      password.DefaultRealm,
				Password:   "123456",
			}
			userID := "john.doe.id"

			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)

			authResp, ok := resp.(model.AuthResponse)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)

			// check the authinfo store data
			a := authinfo.AuthInfo{}
			authInfoStore.GetAuth(userID, &a)
			So(a.LastLoginAt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)
			So(a.LastSeenAt.Equal(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)), ShouldBeTrue)

			// check the token
			tokenStr := authResp.AccessToken
			token := authtoken.Token{}
			tokenStore.Get(tokenStr, &token)
			So(token.AuthInfoID, ShouldEqual, userID)
			So(!token.IsExpired(), ShouldBeTrue)
		})

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
			So(err.Error(), ShouldEqual, "ResourceNotFound: user not found")
		})

		Convey("login user with incorrect password", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Realm:      password.DefaultRealm,
				Password:   "wrong_password",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidCredentials: login_id or password incorrect")
		})

		Convey("login with incorrect login_id", func() {
			payload := LoginRequestPayload{
				LoginIDKey: "phone",
				LoginID:    "202-111-2222",
				Realm:      password.DefaultRealm,
				Password:   "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: login ID key is not allowed")
		})

		Convey("login with disallowed realm", func() {
			payload := LoginRequestPayload{
				LoginID:  "john.doe+1@example.com",
				Realm:    "test",
				Password: "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: realm is not allowed")
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
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
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
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					Realm:          password.DefaultRealm,
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)

		lh := &LoginHandler{}
		lh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		lh.TokenStore = mockTokenStore
		lh.PasswordAuthProvider = passwordAuthProvider
		lh.IdentityProvider = principal.NewMockIdentityProvider(lh.PasswordAuthProvider)
		lh.AuditTrail = coreAudit.NewMockTrail(t)
		profileData := map[string]map[string]interface{}{
			userID: map[string]interface{}{},
		}
		lh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		h := handler.APIHandlerToHandler(lh, lh.TxContext)

		Convey("should contains current identity", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_id_key": "email",
				"login_id": "john.doe@example.com",
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user": {
						"id": "%s",
						"is_verified": true,
						"is_disabled": false,
						"last_login_at": "2006-01-02T15:04:05Z",
						"created_at": "0001-01-01T00:00:00Z",
						"verify_info": {},
						"metadata": {},
						"identity": {
							"id": "john.doe.principal.id1",
							"type": "password",
							"login_id_key": "email",
							"login_id": "john.doe@example.com",
							"realm": "default",
							"claims": {}
						}
					},
					"access_token": "%s"
				}
			}`,
				userID,
				token.AccessToken,
			))
		})
	})
}
