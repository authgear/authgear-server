package push

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	// "log"
	"errors"

	"github.com/anachronistic/apns"
)

type naiveClient struct {
	pn   *apns.PushNotification
	resp *apns.PushNotificationResponse
	apns.Client
}

func (c *naiveClient) Send(pn *apns.PushNotification) *apns.PushNotificationResponse {
	c.pn = pn
	return c.resp
}

func TestSend(t *testing.T) {
	Convey("APNSPusher", t, func() {
		client := naiveClient{}
		pusher := APNSPusher{
			Client: &client,
		}

		Convey("pushes notification", func() {
			client.resp = &apns.PushNotificationResponse{
				Success: true,
			}

			customMap := MapMapper{
				"string":  "value",
				"integer": 1,
				"nested": map[string]interface{}{
					"should": "correct",
				},
			}

			err := pusher.Send(customMap, "deviceToken")

			So(err, ShouldBeNil)

			pn := client.pn
			So(pn.Get("aps"), ShouldResemble, &apns.Payload{
				ContentAvailable: 1,
			})
			So(pn.DeviceToken, ShouldEqual, "deviceToken")
			So(pn.Get("string"), ShouldEqual, "value")
			So(pn.Get("integer"), ShouldEqual, 1)
			So(pn.Get("nested"), ShouldResemble, map[string]interface{}{
				"should": "correct",
			})
		})

		Convey("returns error returned from Client.Send", func() {
			client.resp = &apns.PushNotificationResponse{
				Success: false,
				Error:   errors.New("apns_test: some error"),
			}

			err := pusher.Send(MapMapper{}, "deviceToken")
			So(err, ShouldResemble, errors.New("apns_test: some error"))
		})
	})
}
