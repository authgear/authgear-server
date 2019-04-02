package handler

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestLoginHandler(t *testing.T) {
	Convey("Test LoginRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := LoginRequestPayload{
				RawAuthData: map[string]string{
					"username": "john.doe",
				},
				AuthDataKey: "username",
				AuthData:    "john.doe",
				Password:    "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without auth data", func() {
			payload := LoginRequestPayload{
				Password: "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := LoginRequestPayload{
				RawAuthData: map[string]string{
					"username": "john.doe",
				},
				AuthDataKey: "username",
				AuthData:    "john.doe",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without auth data key", func() {
			payload := LoginRequestPayload{
				RawAuthData: map[string]string{},
				Password:    "123456",
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
					AuthDataKey:    "email",
					AuthData:       "john.doe@example.com",
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id2": password.Principal{
					ID:             "john.doe.principal.id2",
					UserID:         "john.doe.id",
					AuthDataKey:    "username",
					AuthData:       "john.doe",
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

		Convey("login user with auth data", func() {
			authData := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawAuthData: authData,
				AuthDataKey: "email",
				AuthData:    "john.doe@example.com",
				Password:    "123456",
			}
			userID := "john.doe.id"

			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)

			authResp, ok := resp.(response.AuthResponse)
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
			authData := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawAuthData: authData,
				AuthDataKey: "email",
				AuthData:    "john.doe@example.com",
				Password:    "wrong_password",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidCredentials: auth_data or password incorrect")
		})

		Convey("login with incorrect auth data", func() {
			authData := map[string]string{
				"phone": "202-111-2222",
			}
			payload := LoginRequestPayload{
				RawAuthData: authData,
				AuthDataKey: "phone",
				AuthData:    "202-111-2222",
				Password:    "123456",
			}
			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "InvalidArgument: invalid auth data, check your LOGIN_IDS_KEY_WHITELIST setting")
		})

		Convey("log audit trail when login success", func() {
			authData := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawAuthData: authData,
				AuthDataKey: "email",
				AuthData:    "john.doe@example.com",
				Password:    "123456",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_success")
		})

		Convey("log audit trail when login fail", func() {
			authData := map[string]string{
				"email": "john.doe@example.com",
			}
			payload := LoginRequestPayload{
				RawAuthData: authData,
				AuthDataKey: "email",
				AuthData:    "john.doe@example.com",
				Password:    "wrong_password",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "login_failure")
		})
	})
}
