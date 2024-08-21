package idpsession

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

func TestComputeSessionExpiry(t *testing.T) {
	enabled := true
	disabled := false

	Convey("computeSessionExpiry", t, func() {
		session := &IDPSession{
			ID:        "session-id",
			CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			AccessInfo: access.Info{
				LastAccess: access.Event{
					Timestamp: time.Date(2020, 1, 1, 0, 0, 25, 0, time.UTC),
				},
			},
		}

		Convey("idle timeout is disabled", func() {
			setSessionExpireAtForResolvedSession(session, &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &disabled,
				IdleTimeout:        30,
			})
			So(session.ExpireAtForResolvedSession, ShouldResemble, time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC))
		})

		Convey("idle timeout is enabled", func() {
			setSessionExpireAtForResolvedSession(session, &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &enabled,
				IdleTimeout:        30,
			})
			So(session.ExpireAtForResolvedSession, ShouldResemble, time.Date(2020, 1, 1, 0, 0, 55, 0, time.UTC))
		})
	})
}
