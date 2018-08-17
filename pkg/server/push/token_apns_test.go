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

	"github.com/skygeario/buford/push"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

const testTokenKey = `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgtlPE1x3qm2KVqvevMU71bkSvrcN+fDcEDxKCjPtkfzigCgYIKoZIzj0DAQehRANCAARR/TUawEGk9/VzkcAukYYQ2NMfsYbCqU1JBgmJ3pwBQ/BSMRCDXv1qG8uIrUK6Jg494dEWB+RT39sO+vDwgnzD
-----END PRIVATE KEY-----
`

func TestTokenBaseAPNSPusherCreate(t *testing.T) {
	Convey("create token based APNS from token", t, func() {
		pusher, err := NewTokenBasedAPNSPusher(
			nil,
			Sandbox,
			"test-team-id",
			"test-key-id",
			testTokenKey,
		)
		So(err, ShouldBeNil)
		So(pusher, ShouldNotBeNil)
	})
}

func TestTokenBaseAPNSPusherSend(t *testing.T) {
	Convey("TokenBaseAPNSPusher", t, func() {
		conn := &mockConn{}
		service := naiveService{}

		pusher, err := NewTokenBasedAPNSPusher(
			nil,
			Sandbox,
			"test-team-id",
			"test-key-id",
			testTokenKey,
		)
		So(err, ShouldBeNil)

		pusher.(*tokenBasedAPNSPusher).service = &service
		pusher.(*tokenBasedAPNSPusher).connOpener = conn.Open
		pusher.(*tokenBasedAPNSPusher).conn = conn
		pusher.(*tokenBasedAPNSPusher).failed = make(chan failedNotification, 10)
		pusher.(*tokenBasedAPNSPusher).refreshToken()

		device := skydb.Device{
			Token: "deviceToken",
			Topic: "deviceTopic",
		}
		secondDevice := skydb.Device{
			Token: "deviceToken2",
			Topic: "deviceTopic2",
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
			So(service.Sent[0].Headers.Topic, ShouldEqual, "deviceTopic")
			So(service.Sent[0].Headers.Authorization, ShouldNotBeEmpty)
			So(string(service.Sent[0].Payload), ShouldEqualJSON, `{
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

			So(service.Sent[1].DeviceToken, ShouldEqual, "deviceToken2")
			So(service.Sent[1].Headers.Topic, ShouldEqual, "deviceTopic2")
			So(service.Sent[1].Headers.Authorization, ShouldNotBeEmpty)
			So(string(service.Sent[1].Payload), ShouldEqualJSON, `{
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
			So(<-pusher.getFailedNotificationChannel(), ShouldResemble, failedNotification{
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

func TestTokenBaseAPNSPusherFeedbackInterface(t *testing.T) {
	Convey("TokenBaseAPNSPusher", t, func() {
		conn := &mockConn{}
		service := naiveService{}

		pusher, err := NewTokenBasedAPNSPusher(
			nil,
			Sandbox,
			"test-team-id",
			"test-key-id",
			testTokenKey,
		)
		So(err, ShouldBeNil)

		pusher.(*tokenBasedAPNSPusher).service = &service
		pusher.(*tokenBasedAPNSPusher).connOpener = conn.Open
		pusher.(*tokenBasedAPNSPusher).conn = conn
		pusher.(*tokenBasedAPNSPusher).failed = make(chan failedNotification, 10)
		pusher.(*tokenBasedAPNSPusher).refreshToken()

		Convey("contains failed notification channel", func() {
			pusher.Start()
			defer pusher.Stop()

			failed := pusher.getFailedNotificationChannel()
			So(failed, ShouldNotBeNil)
		})

		Convey("can unregister devices", func() {
			pusher.deleteDeviceToken(
				"token-to-be-deleted-1",
				time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			)
			So(len(conn.calls), ShouldEqual, 1)
			So(conn.calls[0].token, ShouldEqual, "token-to-be-deleted-1")
			So(conn.calls[0].t.Unix(), ShouldEqual, 1136214245)
		})
	})
}
