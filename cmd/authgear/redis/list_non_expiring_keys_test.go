package redis

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListNonExpiringKeys(t *testing.T) {
	memoryRedis := miniredis.RunT(t)

	Convey("ListNonExpiringKeys", t, func() {
		ctx := context.Background()
		redisClient := goredis.NewClient(&goredis.Options{Addr: memoryRedis.Addr()})

		// We first setup some keys
		for i := 0; i < 100; i += 1 {
			key := fmt.Sprintf("expiring-%02d", i)
			value := key
			_, err := redisClient.Set(ctx, key, value, 5*time.Second).Result()
			So(err, ShouldBeNil)
		}
		for i := 0; i < 100; i += 1 {
			key := fmt.Sprintf("non-expiring-%02d", i)
			value := key
			_, err := redisClient.Set(ctx, key, value, 0).Result()
			So(err, ShouldBeNil)
		}

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		logger := log.New(stderr, "", 0)
		err := ListNonExpiringKeys(ctx, redisClient, stdout, logger)
		So(err, ShouldBeNil)

		So(stderr.String(), ShouldEqual, `SCAN with cursor 0: 200
done SCAN: 200
`)
		So(stdout.String(), ShouldEqual, `non-expiring-00
non-expiring-01
non-expiring-02
non-expiring-03
non-expiring-04
non-expiring-05
non-expiring-06
non-expiring-07
non-expiring-08
non-expiring-09
non-expiring-10
non-expiring-11
non-expiring-12
non-expiring-13
non-expiring-14
non-expiring-15
non-expiring-16
non-expiring-17
non-expiring-18
non-expiring-19
non-expiring-20
non-expiring-21
non-expiring-22
non-expiring-23
non-expiring-24
non-expiring-25
non-expiring-26
non-expiring-27
non-expiring-28
non-expiring-29
non-expiring-30
non-expiring-31
non-expiring-32
non-expiring-33
non-expiring-34
non-expiring-35
non-expiring-36
non-expiring-37
non-expiring-38
non-expiring-39
non-expiring-40
non-expiring-41
non-expiring-42
non-expiring-43
non-expiring-44
non-expiring-45
non-expiring-46
non-expiring-47
non-expiring-48
non-expiring-49
non-expiring-50
non-expiring-51
non-expiring-52
non-expiring-53
non-expiring-54
non-expiring-55
non-expiring-56
non-expiring-57
non-expiring-58
non-expiring-59
non-expiring-60
non-expiring-61
non-expiring-62
non-expiring-63
non-expiring-64
non-expiring-65
non-expiring-66
non-expiring-67
non-expiring-68
non-expiring-69
non-expiring-70
non-expiring-71
non-expiring-72
non-expiring-73
non-expiring-74
non-expiring-75
non-expiring-76
non-expiring-77
non-expiring-78
non-expiring-79
non-expiring-80
non-expiring-81
non-expiring-82
non-expiring-83
non-expiring-84
non-expiring-85
non-expiring-86
non-expiring-87
non-expiring-88
non-expiring-89
non-expiring-90
non-expiring-91
non-expiring-92
non-expiring-93
non-expiring-94
non-expiring-95
non-expiring-96
non-expiring-97
non-expiring-98
non-expiring-99
`)
	})
}
