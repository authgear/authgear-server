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

package subscription

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
	. "github.com/smartystreets/goconvey/convey"
)

type notifyFunc func(device skydb.Device, notice Notice) error

func (f notifyFunc) CanNotify(device skydb.Device) bool {
	return true
}

func (f notifyFunc) Notify(device skydb.Device, notice Notice) error {
	return f(device, notice)
}

func TestService(t *testing.T) {
	Convey("Subscription Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		timeNow = func() time.Time { return time.Unix(0x43b940e5, 0) }
		defer func() {
			timeNow = time.Now
		}()

		conn := mock_skydb.NewMockConn(ctrl)
		db := mock_skydb.NewMockDatabase(ctrl)

		service := &Service{
			ConnOpener: func() (skydb.Conn, error) { return conn, nil },
		}

		chch := make(chan chan skydb.RecordEvent, 1)
		conn.EXPECT().Subscribe(gomock.Any()).Do(func(recordEventCh chan skydb.RecordEvent) {
			chch <- recordEventCh
		})
		go service.Run()
		defer service.Stop()
		ch := <-chch

		record := skydb.Record{
			ID: skydb.NewRecordID("record", "0"),
		}
		subscription := skydb.Subscription{
			ID:       "subscriptionid",
			DeviceID: "deviceid",
		}
		device := skydb.Device{
			ID: "deviceid",
		}

		conn.EXPECT().PublicDB().Return(db).AnyTimes()
		db.EXPECT().GetMatchingSubscriptions(&record).Return([]skydb.Subscription{
			subscription,
		}).AnyTimes()
		db.EXPECT().Conn().Return(conn).AnyTimes()
		conn.EXPECT().GetDevice("deviceid", gomock.Any()).
			SetArg(1, device).
			Return(nil).
			AnyTimes()

		Convey("sends notice", func() {
			var (
				d skydb.Device
				n Notice
			)
			done := make(chan bool)
			service.Notifier = notifyFunc(func(device skydb.Device, notice Notice) error {
				d = device
				n = notice
				done <- true
				return nil
			})

			ch <- skydb.RecordEvent{
				Record: &record,
				Event:  skydb.RecordCreated,
			}

			select {
			case <-done:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Receive no notices after 100 ms")
			}

			So(d, ShouldResemble, device)
			So(n, ShouldResemble, Notice{
				SeqNum:         0x43b940e50000000,
				SubscriptionID: "subscriptionid",
				Event:          skydb.RecordCreated,
				Record:         &record,
			})
		})

		Convey("increments sequence number", func() {
			var n Notice
			done := make(chan bool)
			service.Notifier = notifyFunc(func(device skydb.Device, notice Notice) error {
				n = notice
				done <- true
				return nil
			})

			ch <- skydb.RecordEvent{Record: &record, Event: skydb.RecordCreated}
			<-done
			So(n.SeqNum, ShouldEqual, 0x43b940e50000000)

			ch <- skydb.RecordEvent{Record: &record, Event: skydb.RecordCreated}
			<-done
			So(n.SeqNum, ShouldEqual, 0x43b940e50000001)
		})

		Convey("resets sequence number on next second", func() {
			var n Notice
			done := make(chan bool)
			service.Notifier = notifyFunc(func(device skydb.Device, notice Notice) error {
				n = notice
				done <- true
				return nil
			})

			ch <- skydb.RecordEvent{Record: &record, Event: skydb.RecordCreated}
			<-done
			So(n.SeqNum, ShouldEqual, 0x43b940e50000000)

			timeNow = func() time.Time { return time.Unix(0x43b940e6, 0) }
			ch <- skydb.RecordEvent{Record: &record, Event: skydb.RecordCreated}
			<-done
			So(n.SeqNum, ShouldEqual, 0x43b940e60000000)
		})
	})
}
