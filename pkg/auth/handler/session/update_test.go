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

func TestUpdateHandler(t *testing.T) {
	Convey("Test UpdateHandler", t, func() {
		h := &UpdateHandler{}
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
			Name:        "Test Session 1",
			CustomData:  nil,
		}
		sessionProvider.Sessions["user-id-1-principal-id-2"] = auth.Session{
			ID:          "user-id-1-principal-id-2",
			ClientID:    "web-app",
			UserID:      "user-id-1",
			PrincipalID: "principal-id-2",
			CreatedAt:   now,
			AccessedAt:  now,
			Name:        "Test Session 2",
			CustomData:  map[string]interface{}{"test": 123},
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

		Convey("should update just name", func() {
			newName := "Name"
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-2",
				Name:      &newName,
				Data:      nil,
			}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].Name, ShouldEqual, "Test Session 1")
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].CustomData, ShouldBeNil)
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].Name, ShouldEqual, "Name")
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].CustomData, ShouldResemble, map[string]interface{}{"test": 123})
		})
		Convey("should clear just name", func() {
			newName := ""
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-2",
				Name:      &newName,
				Data:      nil,
			}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].Name, ShouldEqual, "Test Session 1")
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].CustomData, ShouldBeNil)
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].Name, ShouldEqual, "")
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].CustomData, ShouldResemble, map[string]interface{}{"test": 123})
		})
		Convey("should update just custom data", func() {
			authContext.UseMasterKey()
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-1",
				Name:      nil,
				Data:      map[string]interface{}{"test": 999},
			}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].Name, ShouldEqual, "Test Session 1")
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].CustomData, ShouldResemble, map[string]interface{}{"test": 999})
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].Name, ShouldEqual, "Test Session 2")
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].CustomData, ShouldResemble, map[string]interface{}{"test": 123})
		})
		Convey("should clear just custom data", func() {
			authContext.UseMasterKey()
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-2",
				Name:      nil,
				Data:      map[string]interface{}{},
			}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].Name, ShouldEqual, "Test Session 1")
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].CustomData, ShouldBeNil)
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].Name, ShouldEqual, "Test Session 2")
			So(sessionProvider.Sessions["user-id-1-principal-id-2"].CustomData, ShouldResemble, map[string]interface{}{})
		})
		Convey("should reject custom data update without master key", func() {
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-1",
				Name:      nil,
				Data:      map[string]interface{}{},
			}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "PermissionDenied: must update custom data using master key")
		})
		Convey("should update name & custom data", func() {
			authContext.UseMasterKey()
			newName := "Name"
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-1",
				Name:      &newName,
				Data:      map[string]interface{}{"test": 999},
			}
			resp, err := h.Handle(payload)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, map[string]string{})
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].Name, ShouldEqual, "Name")
			So(sessionProvider.Sessions["user-id-1-principal-id-1"].CustomData, ShouldResemble, map[string]interface{}{"test": 999})
		})
		Convey("should reject non-existing session", func() {
			name := "Test"
			payload := UpdateRequestPayload{
				SessionID: "user-id-1-principal-id-4",
				Name:      &name,
				Data:      nil,
			}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "ResourceNotFound: session not found")
		})
		Convey("should reject session of other users", func() {
			name := "Test"
			payload := UpdateRequestPayload{
				SessionID: "user-id-2-principal-id-2",
				Name:      &name,
				Data:      nil,
			}
			_, err := h.Handle(payload)
			So(err, ShouldBeError, "ResourceNotFound: session not found")
		})
	})
}
