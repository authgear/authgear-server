package whatsapp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	. "github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMessageStore(t *testing.T) {
	Convey("MessageStore", t, func() {
		mr := miniredis.RunT(t)

		client := goredis.NewClient(&goredis.Options{
			Addr: mr.Addr(),
		})
		defer client.Close()

		pool := redis.NewPool()
		So(pool, ShouldNotBeNil)

		connectionOptions := redis.ConnectionOptions{
			RedisURL:          "redis://" + mr.Addr(),
			MaxOpenConnection: func(i int) *int { return &i }(10),
			MaxIdleConnection: func(i int) *int { return &i }(5),
			IdleConnectionTimeout: func(d time.Duration) *config.DurationSeconds {
				ds := config.DurationSeconds(d.Seconds())
				return &ds
			}(duration.UserInteraction),
			MaxConnectionLifetime: func(d time.Duration) *config.DurationSeconds {
				ds := config.DurationSeconds(d.Seconds())
				return &ds
			}(duration.UserInteraction),
		}

		rh := redis.NewHandle(pool, connectionOptions)
		So(rh, ShouldNotBeNil)

		g := &globalredis.Handle{
			Handle: rh,
		}

		s := &MessageStore{
			Redis: g,
			Credentials: &config.WhatsappCloudAPICredentials{
				PhoneNumberID: "1234567890",
			},
		}

		ctx := context.Background()

		Convey("UpdateMessageStatus and GetMessageStatus", func() {
			messageID := "test_message_id"
			statusData := &WhatsappMessageStatusData{Status: WhatsappMessageStatusDelivered}

			err := s.UpdateMessageStatus(ctx, messageID, statusData)
			So(err, ShouldBeNil)

			// Verify data in miniredis
			hashedPhoneNumberID := crypto.SHA256String("1234567890")
			key := fmt.Sprintf("whatsapp:phone-number-id-sha256:%s:message-id:%s", hashedPhoneNumberID, messageID)
			val, err := mr.Get(key)
			So(err, ShouldBeNil)
			expectedbytes, _ := json.Marshal(statusData)
			So(val, ShouldEqual, string(expectedbytes))

			fetchedStatusData, err := s.GetMessageStatus(ctx, messageID)
			So(err, ShouldBeNil)
			So(fetchedStatusData, ShouldResemble, statusData)

			Convey("should return nil for non-existent message", func() {
				fetchedStatusData, err := s.GetMessageStatus(ctx, "non_existent_message_id")
				So(err, ShouldBeNil)
				So(fetchedStatusData, ShouldBeNil)
			})
		})
	})
}
