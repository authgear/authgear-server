package session

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func TestRevokeHandler(t *testing.T) {
	Convey("Test RevokeHandler", t, func() {
		h := &RevokeHandler{}
		h.TxContext = db.NewMockTxContext()
		h.AuthContext = authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1")
		sessionProvider := session.NewMockProvider()
		h.SessionProvider = sessionProvider

		now := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
		sessionProvider.Sessions["user-id-1-principal-id-1"] = auth.Session{
			ID:          "user-id-1-principal-id-1",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-1",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		sessionProvider.Sessions["user-id-1-principal-id-2"] = auth.Session{
			ID:          "user-id-1-principal-id-2",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-2",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		sessionProvider.Sessions["user-id-2-principal-id-3"] = auth.Session{
			ID:          "user-id-2-principal-id-3",
			ClientID:    "web-app",
			UserID:      "user-id-2",
			PrincipalID: "principal-id-3",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		*h.AuthContext.Session() = sessionProvider.Sessions["user-id-1-principal-id-1"]

		Convey("should revoke existing session", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-1-principal-id-2"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldNotContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")
		})

		Convey("should reject current session", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-1-principal-id-1"}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "InvalidArgument: must not revoke current session")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")
		})

		Convey("should ignore non-existing session", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-1-principal-id-4"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
		})

		Convey("should ignore session of other users", func() {
			payload := RevokeRequestPayload{SessionID: "user-id-2-principal-id-3"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-1")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-1-principal-id-2")
			So(sessionProvider.Sessions, ShouldContainKey, "user-id-2-principal-id-3")
		})

	})
}
