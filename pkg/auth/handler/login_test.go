package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
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
				RawLoginID: map[string]string{
					"username": "john.doe",
				},
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
				RawLoginID: map[string]string{
					"username": "john.doe",
				},
				LoginIDKey: "username",
				LoginID:    "john.doe",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without login_id key", func() {
			payload := LoginRequestPayload{
				RawLoginID: map[string]string{},
				Password:   "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
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
		loginIDsKeyWhitelist := []string{"email", "username"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)

		h := &LoginHandler{}
		h.AuthInfoStore = authInfoStore
		h.TokenStore = tokenStore
		h.PasswordAuthProvider = passwordAuthProvider
		h.AuditTrail = coreAudit.NewMockTrail(t)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()

		Convey("login user with login_id", func() {
			loginID := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawLoginID: loginID,
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Password:   "123456",
			}
			userID := "john.doe.id"

			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)

			authResp, ok := resp.(response.User)
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

		Convey("login user with incorrect password", func() {
			loginID := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawLoginID: loginID,
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Password:   "wrong_password",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidCredentials: login_id or password incorrect")
		})

		Convey("login with incorrect login_id", func() {
			loginID := map[string]string{
				"phone": "202-111-2222",
			}
			payload := LoginRequestPayload{
				RawLoginID: loginID,
				LoginIDKey: "phone",
				LoginID:    "202-111-2222",
				Password:   "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: invalid login_id, check your LOGIN_IDS_KEY_WHITELIST setting")
		})

		Convey("log audit trail when login success", func() {
			loginID := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawLoginID: loginID,
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
				Password:   "123456",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_success")
		})

		Convey("log audit trail when login fail", func() {
			loginID := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawLoginID: loginID,
				LoginIDKey: "email",
				LoginID:    "john.doe@example.com",
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
		loginIDsKeyWhitelist := []string{"email", "username"}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDsKeyWhitelist,
			map[string]password.Principal{
				"john.doe.principal.id1": password.Principal{
					ID:             "john.doe.principal.id1",
					UserID:         "john.doe.id",
					LoginIDKey:     "email",
					LoginID:        "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					LoginIDKey:     "username",
					LoginID:        "john.doe",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)

		lh := &LoginHandler{}
		lh.AuthInfoStore = authInfoStore
		mockTokenStore := authtoken.NewMockStore()
		lh.TokenStore = mockTokenStore
		lh.PasswordAuthProvider = passwordAuthProvider
		lh.AuditTrail = coreAudit.NewMockTrail(t)
		profileData := map[string]map[string]interface{}{
			userID: map[string]interface{}{},
		}
		lh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)
		lh.TxContext = db.NewMockTxContext()
		h := handler.APIHandlerToHandler(lh, lh.TxContext)

		Convey("should contains multiple loginIDs", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_id": {
					"email": "john.doe@example.com"
				},
				"password": "123456"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			token := mockTokenStore.GetTokensByAuthInfoID(userID)[0]
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "%s",
					"verified": true,
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
