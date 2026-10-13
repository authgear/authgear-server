package oauth

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestResolverResolveAccessToken(t *testing.T) {
	Convey("Resolver.resolveAccessToken user-match backstop", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		decoder := NewMockAccessTokenDecoder(ctrl)
		accessGrants := NewMockAccessGrantStore(ctrl)
		authzs := NewMockAuthorizationStore(ctrl)
		sessions := NewMockResolverSessionProvider(ctrl)

		re := &Resolver{
			Authorizations:     authzs,
			AccessGrants:       accessGrants,
			AccessTokenDecoder: decoder,
			Sessions:           sessions,
			Clock:              clock.NewMockClockAt("2020-01-01T00:00:00Z"),
		}

		const token = "the-access-token"
		const tokenHash = "the-hash"

		// An IDP-session-backed access grant resolves through the shared
		// post-switch check, so this exercises the backstop for both grant kinds.
		setup := func(grantUserID string, sessionUserID string) {
			decoder.EXPECT().DecodeAccessToken(token).Return(tokenHash, true, nil)
			accessGrants.EXPECT().GetAccessGrant(gomock.Any(), tokenHash).Return(&AccessGrant{
				AuthorizationID: "authz-id",
				UserID:          grantUserID,
				SessionID:       "idp-session-id",
				SessionKind:     GrantSessionKindSession,
			}, nil)
			authzs.EXPECT().GetByID(gomock.Any(), "authz-id").Return(&Authorization{ID: "authz-id"}, nil)
			sessions.EXPECT().AccessWithID(gomock.Any(), "idp-session-id", gomock.Any()).Return(&idpsession.IDPSession{
				ID:    "idp-session-id",
				Attrs: *session.NewAttrs(sessionUserID),
			}, nil)
		}

		Convey("rejects when the bound session belongs to a different user", func() {
			setup("user-a", "user-b")

			s, err := re.resolveAccessToken(context.Background(), token)
			So(s, ShouldBeNil)
			So(errors.Is(err, session.ErrInvalidSession), ShouldBeTrue)
		})

		Convey("resolves when the bound session belongs to the grant's user", func() {
			setup("user-a", "user-a")

			s, err := re.resolveAccessToken(context.Background(), token)
			So(err, ShouldBeNil)
			So(s, ShouldNotBeNil)
			So(s.GetAuthenticationInfo().UserID, ShouldEqual, "user-a")
		})

		Convey("skips the check for legacy grants without a recorded UserID", func() {
			// Grants created before UserID was recorded must still resolve, so
			// the guard only applies when grant.UserID is set.
			setup("", "user-b")

			s, err := re.resolveAccessToken(context.Background(), token)
			So(err, ShouldBeNil)
			So(s, ShouldNotBeNil)
			So(s.GetAuthenticationInfo().UserID, ShouldEqual, "user-b")
		})
	})
}
