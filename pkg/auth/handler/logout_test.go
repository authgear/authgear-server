package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func TestLogoutHandler(t *testing.T) {
	Convey("Test LogoutRequestPayload", t, func() {
		Convey("validate valid payload", func() {
			payload := LogoutRequestPayload{
				AccessToken: "test_token",
			}
			So(payload.Validate(), ShouldBeNil)
		})

		Convey("validate missing access token", func() {
			payload := LogoutRequestPayload{}
			err := payload.Validate()
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.AccessTokenNotAccepted)
		})
	})

	Convey("Test LogoutHandler", t, func() {
		h := &LogoutHandler{}
		h.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		h.TokenStore = authtoken.NewJWTStore("myApp", "secret", 0)
		h.UserProfileStore = userprofile.NewMockUserProfileStore()
		h.AuditTrail = coreAudit.NewMockTrail(t)
		h.HookProvider = hook.NewMockProvider()

		Convey("logout user successfully", func() {
			token := "test_token"
			payload := LogoutRequestPayload{
				AccessToken: token,
			}
			resp, err := h.Handle(payload)
			So(resp, ShouldResemble, map[string]string{})
			So(err, ShouldBeNil)
		})

		Convey("log audit trail when logout", func() {
			payload := LogoutRequestPayload{
				AccessToken: "test_token",
			}
			h.Handle(payload)
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "logout")
		})
	})
}
