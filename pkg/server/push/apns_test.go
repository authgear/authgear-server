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
	. "github.com/smartystreets/goconvey/convey"
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

type deleteCall struct {
	token string
	t     time.Time
}

type mockConn struct {
	calls []deleteCall
	err   error
	skydb.Conn
}

func (c *mockConn) DeleteDevicesByToken(token string, t time.Time) error {
	c.calls = append(c.calls, deleteCall{token, t})
	return c.err
}

func (c *mockConn) Open() (skydb.Conn, error) {
	return c, nil
}

type mockPusher struct {
	APNSPusher

	started          bool
	failed           chan failedNotification
	deleteTokenCalls []deleteCall
}

func (p *mockPusher) Start() {
	p.started = true
}

func (p *mockPusher) Stop() {
	p.started = false
}

func (p mockPusher) getFailedNotificationChannel() chan failedNotification {
	return p.failed
}

func (p *mockPusher) deleteDeviceToken(token string, beforeTime time.Time) error {
	p.deleteTokenCalls = append(p.deleteTokenCalls, deleteCall{
		token: token,
		t:     beforeTime,
	})
	return nil
}

func TestAPNSFeedbackMechanism(t *testing.T) {
	Convey("APNSFeedbackMechanism", t, func() {
		Convey("can queue the failed notifications", func() {
			pusher := &mockPusher{
				failed: make(chan failedNotification, 1),
			}
			queueFailedNotification(pusher, "test-token-1", push.Error{
				Reason:    errors.New("This is a testing error"),
				Status:    http.StatusBadRequest,
				Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})

			failedNoti := <-pusher.failed
			So(failedNoti.deviceToken, ShouldEqual, "test-token-1")
			So(failedNoti.err.Reason.Error(), ShouldEqual, "This is a testing error")
			So(failedNoti.err.Status, ShouldEqual, http.StatusBadRequest)
			So(failedNoti.err.Timestamp.Unix(), ShouldEqual, 1136214245)
		})

		Convey("can trigger unregister device when found bad device token", func() {
			pusher := &mockPusher{
				failed:           make(chan failedNotification, 1),
				deleteTokenCalls: []deleteCall{},
			}

			go checkFailedNotifications(pusher)
			queueFailedNotification(pusher, "test-token-2", push.Error{
				Reason:    errors.New("BadDeviceToken"),
				Status:    http.StatusGone,
				Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})

			time.Sleep(500 * time.Millisecond)
			So(len(pusher.deleteTokenCalls), ShouldEqual, 1)
			So(pusher.deleteTokenCalls[0].token, ShouldEqual, "test-token-2")
			So(pusher.deleteTokenCalls[0].t.Unix(), ShouldEqual, 1136214245)
		})
	})
}
