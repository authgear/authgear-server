package session

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func TestGetHandler(t *testing.T) {
	Convey("Test GetHandler", t, func() {
		h := &GetHandler{}
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
		sessionProvider.Sessions["user-id-2-principal-id-2"] = auth.Session{
			ID:          "user-id-2-principal-id-2",
			ClientID:    "web-app",
			UserID:      "user-id-2",
			PrincipalID: "principal-id-2",
			CreatedAt:   now,
			AccessedAt:  now,
		}
		*h.AuthContext.Session() = sessionProvider.Sessions["user-id-1-principal-id-1"]

		Convey("should get existing session", func() {
			payload := GetRequestPayload{SessionID: "user-id-1-principal-id-1"}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, GetResponse{
				Session: model.Session{
					ID:             "user-id-1-principal-id-1",
					IdentityID:     "principal-id-1",
					CreatedAt:      now,
					LastAccessedAt: now,
					Data:           map[string]interface{}{},
				},
			})
		})

		Convey("should reject non-existing session", func() {
			payload := GetRequestPayload{SessionID: "user-id-1-principal-id-2"}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "ResourceNotFound: session not found")
		})

		Convey("should reject session of other users", func() {
			payload := GetRequestPayload{SessionID: "user-id-2-principal-id-2"}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "ResourceNotFound: session not found")
		})

	})
}
