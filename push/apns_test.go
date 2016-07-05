// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package push

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/RobotsAndPencils/buford/push"
	"github.com/skygeario/skygear-server/skydb"
	. "github.com/skygeario/skygear-server/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timehop/apns"
)

type naiveServiceNotification struct {
	DeviceToken string
	Headers     *push.Headers
	Payload     []byte
}

type naiveService struct {
	Sent []naiveServiceNotification
	Err  *push.Error
}

func (s *naiveService) Push(deviceToken string, headers *push.Headers, payload []byte) (string, error) {
	s.Sent = append(s.Sent, naiveServiceNotification{deviceToken, headers, payload})
	if s.Err != nil {
		return "", s.Err
	}
	return "77BAF428-6FD8-42DB-8D1E-6F14A36C0863", nil
}

func TestAPNSSend(t *testing.T) {
	Convey("APNSPusher", t, func() {
		service := naiveService{}
		pusher := APNSPusher{
			service: &service,
			failed:  make(chan failedNotification, 10),
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
			So(len(service.Sent), ShouldEqual, 2)
			So(service.Sent[0].DeviceToken, ShouldEqual, "deviceToken")
			So(service.Sent[1].DeviceToken, ShouldEqual, "deviceToken2")

			for i := range service.Sent {
				n := service.Sent[i]
				So(string(n.Payload), ShouldEqualJSON, `{
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

		Convey("returns error returned from Service.Push (BadMessageId)", func() {
			service.Err = &push.Error{
				Reason:    errors.New("BadMessageId"),
				Status:    http.StatusBadRequest,
				Timestamp: time.Time{},
			}
			err := pusher.Send(MapMapper{
				"apns": map[string]interface{}{},
			}, device)
			So(err, ShouldResemble, service.Err)
		})

		Convey("returns error returned from Service.Push (Unregistered)", func() {
			pushError := push.Error{
				Reason:    errors.New("Unregistered"),
				Status:    http.StatusGone,
				Timestamp: time.Now(),
			}
			service.Err = &pushError
			err := pusher.Send(MapMapper{
				"apns": map[string]interface{}{},
			}, device)
			So(err, ShouldResemble, &pushError)
			So(<-pusher.failed, ShouldResemble, failedNotification{
				deviceToken: device.Token,
				err:         pushError,
			})
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

			n := service.Sent[0]
			So(n.DeviceToken, ShouldEqual, "deviceToken")

			So(string(n.Payload), ShouldEqualJSON, `{
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
		pusher := APNSPusher{
			connOpener: conn.Open,
			conn:       conn,
		}

		Convey("unregister device", func() {
			pusher.unregisterDevice("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("unregister device with error", func() {
			conn.err = errors.New("apns/test: unknown error")
			pusher.unregisterDevice("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("handle unregistered notification", func() {
			pusher.handleFailedNotification(failedNotification{"devicetoken0", push.Error{
				Reason:    errors.New("Unregistered"),
				Status:    http.StatusGone,
				Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}})
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("check failed notifications", func() {
			pusher.failed = make(chan failedNotification)
			go func() {
				pushError1 := push.Error{
					Reason:    errors.New("Unregistered"),
					Status:    http.StatusGone,
					Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				}
				pusher.failed <- failedNotification{"devicetoken0", pushError1}

				pushError2 := push.Error{
					Reason:    errors.New("Unregistered"),
					Status:    http.StatusGone,
					Timestamp: time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC),
				}
				pusher.failed <- failedNotification{"devicetoken1", pushError2}
				close(pusher.failed)
			}()

			pusher.checkFailedNotifications()
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
				{"devicetoken1", time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})
	})
}
