package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
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
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			map[string]config.LoginIDKeyConfiguration{},
			[]string{password.DefaultRealm},
			map[string]password.Principal{
				"faseng.cat.principal.id": password.Principal{
					ID:         "faseng.cat.principal.id",
					UserID:     "faseng.cat.id",
					LoginIDKey: "email",
					LoginID:    "faseng@example.com",
					Realm:      password.DefaultRealm,
				},
			},
		)
		h.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider)
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("logout user successfully", func() {
			token := "test_token"
			payload := LogoutRequestPayload{
				AccessToken: token,
			}
			resp, err := h.Handle(payload)
			So(resp, ShouldResemble, map[string]string{})
			So(err, ShouldBeNil)

			So(hookProvider.DispatchedEvents, ShouldResemble, []event.Payload{
				event.SessionDeleteEvent{
					Reason: event.SessionDeleteReasonLogout,
					User: model.User{
						ID:         "faseng.cat.id",
						Verified:   true,
						VerifyInfo: map[string]bool{},
						Metadata:   userprofile.Data{},
					},
					Identity: model.Identity{
						ID:   "faseng.cat.principal.id",
						Type: "password",
						Attributes: principal.Attributes{
							"login_id_key": "email",
							"login_id":     "faseng@example.com",
							"realm":        "default",
						},
						Claims: principal.Claims{},
					},
				},
			})
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
