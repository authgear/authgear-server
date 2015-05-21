package push

import (
	"encoding/json"
	"errors"
	. "github.com/oursky/ourd/ourtest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timehop/apns"
	"testing"
)

type naiveClient struct {
	failedNotifs     chan apns.NotificationResult
	lastnotification apns.Notification
	returnerr        error
}

func (c *naiveClient) Send(n apns.Notification) error {
	c.lastnotification = n
	return c.returnerr
}

func (c *naiveClient) FailedNotifs() chan apns.NotificationResult {
	return c.failedNotifs
}

func TestSend(t *testing.T) {
	Convey("APNSPusher", t, func() {
		client := naiveClient{}
		pusher := APNSPusher{
			Client: &client,
		}

		Convey("pushes notification", func() {
			customMap := MapMapper{
				"string":  "value",
				"integer": 1,
				"nested": map[string]interface{}{
					"should": "correct",
				},
			}

			err := pusher.Send(customMap, "deviceToken")

			So(err, ShouldBeNil)

			n := client.lastnotification
			So(n.DeviceToken, ShouldEqual, "deviceToken")

			payloadJSON, _ := json.Marshal(&n.Payload)
			So(payloadJSON, ShouldEqualJSON, `{
	"aps": {
		"content-available": 1
	},
	"string": "value",
	"integer": 1,
	"nested": {
		"should": "correct"
	}
}`)
		})

		Convey("returns error returned from Client.Send", func() {
			client.returnerr = errors.New("apns_test: some error")
			err := pusher.Send(MapMapper{}, "deviceToken")
			So(err, ShouldResemble, errors.New("apns_test: some error"))
		})
	})
}
