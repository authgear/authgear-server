package push

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	. "github.com/oursky/skygear/ourtest"
	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timehop/apns"
)

type naiveClient struct {
	failedNotifs      chan apns.NotificationResult
	sentNotifications []apns.Notification
	returnerr         error
}

func (c *naiveClient) Send(n apns.Notification) error {
	c.sentNotifications = append(c.sentNotifications, n)
	return c.returnerr
}

func (c *naiveClient) FailedNotifs() chan apns.NotificationResult {
	return c.failedNotifs
}

func TestAPNSSend(t *testing.T) {
	Convey("APNSPusher", t, func() {
		client := naiveClient{}
		pusher := APNSPusher{
			client: &client,
		}
		device := skydb.Device{
			Token: "deviceToken",
		}
		secondDevice := skydb.Device{
			Token: "deviceToken2",
		}

		Convey("pushes notification", func() {
			customMap := MapMapper{
				"apns": map[string]interface{}{
					"aps": map[string]interface{}{
						"content-available": 1,
						"sound":             "sosumi.mp3",
						"badge":             5,
						"alert":             "This is a message.",
					},
					"string":  "value",
					"integer": 1,
					"nested": map[string]interface{}{
						"should": "correct",
					},
				},
			}

			So(pusher.Send(customMap, device), ShouldBeNil)
			So(pusher.Send(customMap, secondDevice), ShouldBeNil)
			So(len(client.sentNotifications), ShouldEqual, 2)
			So(client.sentNotifications[0].DeviceToken, ShouldEqual, "deviceToken")
			So(client.sentNotifications[1].DeviceToken, ShouldEqual, "deviceToken2")

			for i := range client.sentNotifications {
				n := client.sentNotifications[i]
				payloadJSON, _ := json.Marshal(&n.Payload)
				So(payloadJSON, ShouldEqualJSON, `{
					"aps": {
						"content-available": 1,
						"sound": "sosumi.mp3",
						"badge": 5,
						"alert": "This is a message."
					},
					"string": "value",
					"integer": 1,
					"nested": {
						"should": "correct"
					}
				}`)
			}
		})

		Convey("returns error when missing apns dictionary", func() {
			err := pusher.Send(EmptyMapper, device)
			So(err, ShouldResemble, errors.New("push/apns: payload has no apns dictionary"))
		})

		Convey("returns error returned from Client.Send", func() {
			client.returnerr = errors.New("apns_test: some error")
			err := pusher.Send(MapMapper{
				"apns": map[string]interface{}{},
			}, device)
			So(err, ShouldResemble, errors.New("apns_test: some error"))
		})

		Convey("pushes with custom alert", func() {
			customMap := MapMapper{
				"apns": map[string]interface{}{
					"aps": map[string]interface{}{
						"alert": map[string]interface{}{
							"body":           "Acme message received from Johnny Appleseed",
							"action-loc-key": "VIEW",
						},
					},
				},
			}

			err := pusher.Send(customMap, device)

			So(err, ShouldBeNil)

			n := client.sentNotifications[0]
			So(n.DeviceToken, ShouldEqual, "deviceToken")

			payloadJSON, _ := json.Marshal(&n.Payload)
			So(payloadJSON, ShouldEqualJSON, `{
				"aps": {
					"alert": {
						"body": "Acme message received from Johnny Appleseed",
						"action-loc-key": "VIEW"
					}
				}
			}`)
		})
	})
}

type deleteCall struct {
	token string
	t     time.Time
}

type mockConn struct {
	calls []deleteCall
	err   error
	skydb.Conn
}

func (c *mockConn) DeleteDeviceByToken(token string, t time.Time) error {
	c.calls = append(c.calls, deleteCall{token, t})
	return c.err
}

func (c *mockConn) Open() (skydb.Conn, error) {
	return c, nil
}

type feedbackChannel chan apns.FeedbackTuple

func (ch feedbackChannel) Receive() <-chan apns.FeedbackTuple {
	return ch
}

func TestAPNSFeedback(t *testing.T) {
	Convey("APNSPusher", t, func() {
		conn := &mockConn{}
		ch := make(chan apns.FeedbackTuple)
		pusher := APNSPusher{
			connOpener: conn.Open,
			feedback:   feedbackChannel(ch),
		}

		Convey("receives no feedbacks", func() {
			close(ch)
			pusher.recvFeedback()
			So(conn.calls, ShouldBeEmpty)
		})

		Convey("receives multiple feedbacks", func() {
			go func() {
				ch <- newFeedbackTuple("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
				ch <- newFeedbackTuple("devicetoken1", time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC))
				close(ch)
			}()

			pusher.recvFeedback()
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
				{"devicetoken1", time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("handles erroneous device delete", func() {
			go func() {
				ch <- newFeedbackTuple("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
				close(ch)
			}()

			conn.err = errors.New("apns/test: unknown error")
			pusher.recvFeedback()
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})
	})
}

func newFeedbackTuple(token string, t time.Time) apns.FeedbackTuple {
	return apns.FeedbackTuple{Timestamp: t, DeviceToken: token}
}
