package lockout

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestAdminStorage(t *testing.T) {
	s := miniredis.RunT(t)

	Convey("getStatus", t, func() {
		ctx := context.Background()

		Convey("PerUser", func() {
			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn()
			defer conn.Close()
			s.FlushAll()

			Convey("not locked when empty", func() {
				key := "test:lockout"
				status, err := getStatus(ctx, conn, key, true)
				So(err, ShouldBeNil)
				So(status.IsLocked, ShouldBeFalse)
				So(status.LockedUntil, ShouldBeNil)
			})

			Convey("locked when lock key exists with future time", func() {
				s.FlushAll()
				key := "test:lockout"
				lockKey := key + ":lock:global"
				futureTime := time.Now().Add(1 * time.Hour).Unix()
				conn.Set(ctx, lockKey, futureTime, 0)

				status, err := getStatus(ctx, conn, key, true)
				So(err, ShouldBeNil)
				So(status.IsLocked, ShouldBeTrue)
				So(status.LockedUntil, ShouldNotBeNil)
			})

			Convey("not locked when lock key is expired", func() {
				s.FlushAll()
				key := "test:lockout"
				lockKey := key + ":lock:global"
				pastTime := int64(1000)
				conn.Set(ctx, lockKey, pastTime, 0)

				status, err := getStatus(ctx, conn, key, true)
				So(err, ShouldBeNil)
				So(status.IsLocked, ShouldBeFalse)
			})
		})

		Convey("PerUserPerIP", func() {
			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn()
			defer conn.Close()

			Convey("not locked when empty", func() {
				s.FlushAll()
				key := "test:lockout"
				status, err := getStatus(ctx, conn, key, false)
				So(err, ShouldBeNil)
				So(status.IsLocked, ShouldBeFalse)
				So(len(status.LockedIPs), ShouldEqual, 0)
			})

			Convey("returns locked IPs with future lock times", func() {
				s.FlushAll()
				key := "test:lockout"
				futureTime := time.Now().Add(1 * time.Hour).Unix()

				lockKey1 := key + ":lock:192.168.1.1"
				conn.Set(ctx, lockKey1, futureTime, 0)

				conn.HSet(ctx, key, "192.168.1.1", 5)
				conn.HSet(ctx, key, "192.168.1.2", 2)

				status, err := getStatus(ctx, conn, key, false)
				So(err, ShouldBeNil)
				So(status.IsLocked, ShouldBeTrue)
				So(len(status.LockedIPs), ShouldEqual, 1)
				So(status.LockedIPs[0].IPAddress, ShouldEqual, "192.168.1.1")
			})
		})
	})

	Convey("clearAll", t, func() {
		ctx := context.Background()

		Convey("PerUser", func() {
			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn()
			defer conn.Close()
			s.FlushAll()

			key := "test:lockout"
			lockKey := key + ":lock:global"
			futureTime := time.Now().Add(1 * time.Hour).Unix()

			conn.Set(ctx, lockKey, futureTime, 0)
			conn.HSet(ctx, key, "total", 5)

			err := clearAll(ctx, conn, key, true)
			So(err, ShouldBeNil)

			status, err := getStatus(ctx, conn, key, true)
			So(err, ShouldBeNil)
			So(status.IsLocked, ShouldBeFalse)
		})

		Convey("PerUserPerIP", func() {
			cli := goredis.NewClient(&goredis.Options{Addr: s.Addr()})
			conn := cli.Conn()
			defer conn.Close()
			s.FlushAll()

			key := "test:lockout"
			futureTime := time.Now().Add(1 * time.Hour).Unix()

			conn.Set(ctx, key+":lock:192.168.1.1", futureTime, 0)
			conn.Set(ctx, key+":lock:192.168.1.2", futureTime, 0)

			conn.HSet(ctx, key, "192.168.1.1", 5)
			conn.HSet(ctx, key, "192.168.1.2", 5)

			err := clearAll(ctx, conn, key, false)
			So(err, ShouldBeNil)

			status, err := getStatus(ctx, conn, key, false)
			So(err, ShouldBeNil)
			So(status.IsLocked, ShouldBeFalse)
			So(len(status.LockedIPs), ShouldEqual, 0)
		})
	})
}
