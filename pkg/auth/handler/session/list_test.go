package session

// TODO(authn): use new session provider
/*
import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	sessiontesting "github.com/skygeario/skygear-server/pkg/auth/dependency/session/testing"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func TestListHandler(t *testing.T) {
	Convey("Test GetRequestPayload", t, func() {
		h := &ListHandler{}
		h.TxContext = db.NewMockTxContext()
		sessionProvider := session.NewMockProvider()
		h.SessionProvider = sessionProvider

		now := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
		sessionProvider.Sessions["user-id-1-principal-id-1"] = auth.Session{
			ID:          "user-id-1-principal-id-1",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-1",
			CreatedAt:   now.Add(-1 * time.Minute),
			AccessedAt:  now.Add(-1 * time.Minute),
		}
		sessionProvider.Sessions["user-id-1-principal-id-2"] = auth.Session{
			ID:          "user-id-1-principal-id-2",
			ClientID:    "mobile-app",
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

		Convey("should list sessions", func() {
			r, _ := http.NewRequest("POST", "/", nil)
			r = sessiontesting.WithSession(r, "user-id-1", "principal-id-1")
			resp, err := h.Handle(r)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, ListResponse{
				Sessions: []model.Session{
					model.Session{
						ID:             "user-id-1-principal-id-1",
						IdentityID:     "principal-id-1",
						CreatedAt:      now.Add(-1 * time.Minute),
						LastAccessedAt: now.Add(-1 * time.Minute),
					},
					model.Session{
						ID:             "user-id-1-principal-id-2",
						IdentityID:     "principal-id-2",
						CreatedAt:      now,
						LastAccessedAt: now,
					},
				},
			})
		})
	})
}
*/
