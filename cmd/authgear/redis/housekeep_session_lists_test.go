package redis

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHousekeepTaskSessionList(t *testing.T) {
	Convey("session-list, whose field is the session key", t, func() {
		f := newHousekeepFixture(t)
		f.set("app:accounts:session:live", "{}", time.Hour)
		f.set("app:accounts:session:stale-expiry", "{}", time.Hour)

		// Deleted: every entry is dead.
		f.hset("app:accounts:session-list:all-dead", "app:accounts:session:dead-1", t_expired)
		f.hset("app:accounts:session-list:all-dead", "app:accounts:session:dead-2", t_expired)

		// Pruned: only the dead entry is removed.
		f.hset("app:accounts:session-list:mixed", "app:accounts:session:dead-3", t_expired)
		// Kept: the session is live.
		f.hset("app:accounts:session-list:mixed", "app:accounts:session:live", t_future)
		// Kept: the application writes the entry before the session key, and the expiry is in the future.
		f.hset("app:accounts:session-list:mixed", "app:accounts:session:being-created", t_future)
		// Kept: the session key still exists, though the stored expiry has passed.
		f.hset("app:accounts:session-list:mixed", "app:accounts:session:stale-expiry", t_expired)
		// Kept: the stored expiry cannot be read.
		f.hsetRaw("app:accounts:session-list:mixed", "app:accounts:session:unreadable", "not-a-time")

		// Left alone: no dead entry.
		f.hset("app:accounts:session-list:only-live", "app:accounts:session:live", t_future)
		// Left alone: not a hash.
		f.set("app:accounts:session-list:not-a-hash", "value", 0)

		stdout, _ := f.run(false, housekeepTaskSessionList)
		So(stdout, ShouldEqual, `deleted app:accounts:session-list:all-dead
pruned app:accounts:session-list:mixed
`)
		So(f.memoryRedis.Keys(), ShouldResemble, []string{
			"app:accounts:session-list:mixed",
			"app:accounts:session-list:not-a-hash",
			"app:accounts:session-list:only-live",
			"app:accounts:session:live",
			"app:accounts:session:stale-expiry",
		})
		So(f.hkeys("app:accounts:session-list:mixed"), ShouldResemble, []string{
			"app:accounts:session:being-created",
			"app:accounts:session:live",
			"app:accounts:session:stale-expiry",
			"app:accounts:session:unreadable",
		})
		So(f.hkeys("app:accounts:session-list:only-live"), ShouldResemble, []string{"app:accounts:session:live"})
	})
}

func TestHousekeepTaskOfflineGrantList(t *testing.T) {
	Convey("offline-grant-list, whose field is the bare offline grant ID", t, func() {
		f := newHousekeepFixture(t)
		f.set("app:accounts:offline-grant:grant-live", "{}", time.Hour)
		f.set("app:accounts:offline-grant:grant-stale-expiry", "{}", time.Hour)
		// The grant of the same ID in another app does not keep the entry.
		f.set("app:other:offline-grant:grant-dead-3", "{}", time.Hour)

		// Deleted: every entry is dead.
		f.hset("app:accounts:offline-grant-list:all-dead", "grant-dead-1", t_expired)
		f.hset("app:accounts:offline-grant-list:all-dead", "grant-dead-2", t_expired)

		// Pruned: only the dead entry is removed.
		f.hset("app:accounts:offline-grant-list:mixed", "grant-dead-3", t_expired)
		// Kept: the grant is live.
		f.hset("app:accounts:offline-grant-list:mixed", "grant-live", t_future)
		// Kept: the application writes the entry before the grant key, and the expiry is in the future.
		f.hset("app:accounts:offline-grant-list:mixed", "grant-being-created", t_future)
		// Kept: the grant key still exists, though the stored expiry has passed.
		f.hset("app:accounts:offline-grant-list:mixed", "grant-stale-expiry", t_expired)
		// Kept: the stored expiry cannot be read.
		f.hsetRaw("app:accounts:offline-grant-list:mixed", "grant-unreadable", "not-a-time")

		// Left alone: no dead entry.
		f.hset("app:accounts:offline-grant-list:only-live", "grant-live", t_future)
		// Left alone: not a hash.
		f.set("app:accounts:offline-grant-list:not-a-hash", "value", 0)

		stdout, _ := f.run(false, housekeepTaskOfflineGrantList)
		So(stdout, ShouldEqual, `deleted app:accounts:offline-grant-list:all-dead
pruned app:accounts:offline-grant-list:mixed
`)
		So(f.memoryRedis.Keys(), ShouldResemble, []string{
			"app:accounts:offline-grant-list:mixed",
			"app:accounts:offline-grant-list:not-a-hash",
			"app:accounts:offline-grant-list:only-live",
			"app:accounts:offline-grant:grant-live",
			"app:accounts:offline-grant:grant-stale-expiry",
			"app:other:offline-grant:grant-dead-3",
		})
		So(f.hkeys("app:accounts:offline-grant-list:mixed"), ShouldResemble, []string{
			"grant-being-created",
			"grant-live",
			"grant-stale-expiry",
			"grant-unreadable",
		})
		So(f.hkeys("app:accounts:offline-grant-list:only-live"), ShouldResemble, []string{"grant-live"})
	})
}
