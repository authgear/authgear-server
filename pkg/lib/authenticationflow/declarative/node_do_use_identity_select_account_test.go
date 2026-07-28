package declarative

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func TestResolveSelectAccountSession(t *testing.T) {
	Convey("resolveSelectAccountSession", t, func() {
		deps := &authflow.Dependencies{HTTPRequest: &http.Request{Header: http.Header{}}}
		makeCtx := func(resolvedSession session.ResolvedSession) context.Context {
			ctx := context.Background()
			if resolvedSession != nil {
				ctx = session.WithSession(ctx, resolvedSession)
			}
			s := &authflow.Session{FlowID: "flow-1"}
			return s.MakeContext(ctx, deps)
		}

		Convey("nil session: SelectAccountSessionChanged", func() {
			ctx := makeCtx(nil)
			userID, sessionID, sessionType, err := resolveSelectAccountSession(ctx, "user-1")
			So(userID, ShouldEqual, "")
			So(sessionID, ShouldEqual, "")
			So(sessionType, ShouldEqual, session.Type(""))
			So(err, ShouldEqual, ErrSelectAccountSessionChanged)
		})

		Convey("mismatched user: SelectAccountSessionChanged", func() {
			ctx := makeCtx(&fakeResolvedSessionForSelectAccount{UserID: "user-2"})
			userID, sessionID, sessionType, err := resolveSelectAccountSession(ctx, "user-1")
			So(userID, ShouldEqual, "")
			So(sessionID, ShouldEqual, "")
			So(sessionType, ShouldEqual, session.Type(""))
			So(err, ShouldEqual, ErrSelectAccountSessionChanged)
		})

		Convey("matching user but session lacks pre-authenticated-url scope: SelectAccountSessionChanged", func() {
			ctx := makeCtx(fakeOfflineGrantSessionWithoutScope("user-1"))
			userID, sessionID, sessionType, err := resolveSelectAccountSession(ctx, "user-1")
			So(userID, ShouldEqual, "")
			So(sessionID, ShouldEqual, "")
			So(sessionType, ShouldEqual, session.Type(""))
			So(err, ShouldEqual, ErrSelectAccountSessionChanged)
		})

		Convey("matching user: no error", func() {
			ctx := makeCtx(&fakeResolvedSessionForSelectAccount{UserID: "user-1"})
			userID, sessionID, sessionType, err := resolveSelectAccountSession(ctx, "user-1")
			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "user-1")
			So(sessionID, ShouldEqual, "session-id")
			So(sessionType, ShouldEqual, session.TypeIdentityProvider)
		})
	})
}
