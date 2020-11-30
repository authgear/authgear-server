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
			expiry := computeSessionStorageExpiry(session, &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &disabled,
				IdleTimeout:        30,
			})
			So(expiry, ShouldResemble, time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC))
		})

		Convey("idle timeout is enabled", func() {
			expiry := computeSessionStorageExpiry(session, &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &enabled,
				IdleTimeout:        30,
			})
			So(expiry, ShouldResemble, time.Date(2020, 1, 1, 0, 0, 55, 0, time.UTC))
		})
	})
}

func TestCheckSessionExpired(t *testing.T) {
	enabled := true
	disabled := false

	Convey("checkSessionExpired", t, func() {
		session := &IDPSession{
			ID:        "session-id",
			CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			AccessInfo: access.Info{
				LastAccess: access.Event{
					Timestamp: time.Date(2020, 1, 1, 0, 0, 25, 0, time.UTC),
				},
			},
		}
		var cfg *config.SessionConfig
		check := func(mins, secs int) bool {
			return !checkSessionExpired(session, time.Date(2020, 1, 1, 0, mins, secs, 0, time.UTC), cfg)
		}

		Convey("check session lifetime", func() {
			cfg = &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &disabled,
				IdleTimeout:        30,
			}

			So(check(0, 0), ShouldBeTrue)
			So(check(0, 56), ShouldBeTrue)
			So(check(2, 0), ShouldBeTrue)
			So(check(2, 1), ShouldBeFalse)
		})

		Convey("check idle timeout", func() {
			cfg = &config.SessionConfig{
				Lifetime:           120,
				IdleTimeoutEnabled: &enabled,
				IdleTimeout:        30,
			}

			So(check(0, 0), ShouldBeTrue)
			So(check(0, 55), ShouldBeTrue)
			So(check(0, 56), ShouldBeFalse)
			So(check(2, 1), ShouldBeFalse)
		})
	})
}
