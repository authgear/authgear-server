package handler

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResetPasswordPayload(t *testing.T) {
	Convey("Test ResetPasswordRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := ResetPasswordRequestPayload{
				AuthInfoID: "1",
				Password:   "123456",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate payload without auth_id", func() {
			payload := ResetPasswordRequestPayload{
				Password: "123456",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("validate payload without password", func() {
			payload := ResetPasswordRequestPayload{
				AuthInfoID: "1",
			}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}
func TestResetPasswordHandler(t *testing.T) {
	Convey("Test ResetPasswordHandler", t, func() {
		// fixture
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"john.doe.id": authinfo.AuthInfo{
					ID: "john.doe.id",
				},
			},
		)
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			map[string]password.Principal{
				"john.doe.principal.id": password.Principal{
					ID:     "john.doe.principal.id",
					UserID: "john.doe.id",
					AuthData: map[string]interface{}{
						"username": "john.doe",
						"email":    "john.doe@example.com",
					},
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		tokenStore := authtoken.NewJWTStore("myApp", "secret", 0)
		passwordChecker := &audit.PasswordChecker{
			PwMinLength: 6,
		}

		h := &ResetPasswordHandler{}
		h.AuthInfoStore = authInfoStore
		h.TokenStore = tokenStore
		h.PasswordChecker = passwordChecker
		h.PasswordAuthProvider = passwordAuthProvider
		h.AuditTrail = coreAudit.NewMockTrail(t)
		h.UserProfileStore = dependency.NewMockUserProfileStore()

		Convey("should reset password by user id", func() {
			userID := "john.doe.id"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "123456",
			}

			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)

			authResp, ok := resp.(response.AuthResponse)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)

			// check the token
			tokenStr := authResp.AccessToken
			token := authtoken.Token{}
			tokenStore.Get(tokenStr, &token)
			So(token.AuthInfoID, ShouldEqual, userID)
			So(!token.IsExpired(), ShouldBeTrue)
		})

		Convey("should not reset password by wrong user id", func() {
			userID := "john.doe.id.wrong"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "123456",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "ResourceNotFound: User not found")
		})

		Convey("should not reset password with password violates password policy", func() {
			userID := "john.doe.id"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "1234",
			}

			_, err := h.Handle(payload)
			So(err.Error(), ShouldEqual, "PasswordPolicyViolated: password too short")
		})

		Convey("should have audit trail when reset password", func() {
			userID := "john.doe.id"
			payload := ResetPasswordRequestPayload{
				AuthInfoID: userID,
				Password:   "123456",
			}

			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "reset_password")
		})
	})
}
