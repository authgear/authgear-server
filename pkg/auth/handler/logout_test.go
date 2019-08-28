package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestLogoutHandler(t *testing.T) {
	Convey("Test LogoutHandler", t, func() {
		h := &LogoutHandler{}

		authContext := authtest.NewMockContext().
			UseUser("faseng.cat.id", "faseng.cat.principal.id").
			MarkVerified()
		authContext.Session().ID = "session-id"
		h.AuthContext = authContext

		sessionProvider := session.NewMockProvider()
		sessionProvider.Sessions[authContext.Session().ID] = *authContext.Session()
		h.SessionProvider = sessionProvider
		h.SessionWriter = authSession.NewMockWriter()

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
					ClaimsValue: map[string]interface{}{
						"email": "faseng@example.com",
					},
				},
			},
		)
		h.IdentityProvider = principal.NewMockIdentityProvider(passwordAuthProvider)
		hookProvider := hook.NewMockProvider()
		h.HookProvider = hookProvider

		Convey("logout user successfully", func() {
			resp, err := h.Handle()
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
						Claims: principal.Claims{
							"email": "faseng@example.com",
						},
					},
				},
			})
		})

		Convey("log audit trail when logout", func() {
			h.Handle()
			mockTrail, _ := h.AuditTrail.(*coreAudit.MockTrail)
			So(mockTrail.Hook.LastEntry().Message, ShouldEqual, "audit_trail")
			So(mockTrail.Hook.LastEntry().Data["event"], ShouldEqual, "logout")
		})
	})
}
