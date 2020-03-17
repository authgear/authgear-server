package session

// TODO(authn): use new session provider
/*
import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authtest "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestGetHandler(t *testing.T) {
	Convey("Test GetHandler", t, func() {
		h := &GetHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			GetRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		authContext := authtest.NewMockContext().
			UseUser("user-id-1", "principal-id-1")
		h.AuthContext = authContext
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
		sess := sessionProvider.Sessions["user-id-1-principal-id-1"]
		authContext.UseSession(&sess)

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
				},
			})
		})

		Convey("should reject non-existing session", func() {
			payload := GetRequestPayload{SessionID: "user-id-1-principal-id-2"}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "session not found")
		})

		Convey("should reject session of other users", func() {
			payload := GetRequestPayload{SessionID: "user-id-2-principal-id-2"}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "session not found")
		})

	})
}
*/
