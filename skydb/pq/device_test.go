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

package pq

import (
	"testing"
	"time"

	"github.com/skygeario/skygear-server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDevice(t *testing.T) {
	Convey("Conn", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		addUser(t, c, "userid")

		Convey("gets an existing Device", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			device = skydb.Device{}
			err := c.GetDevice("deviceid", &device)
			So(err, ShouldBeNil)
			So(device, ShouldResemble, skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("creates a new Device", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldBeNil)

			var (
				deviceType, token, userInfoID string
				lastRegisteredAt              time.Time
			)
			err = c.QueryRowx("SELECT type, token, user_id, last_registered_at FROM app_com_oursky_skygear._device WHERE id = 'deviceid'").
				Scan(&deviceType, &token, &userInfoID, &lastRegisteredAt)
			So(err, ShouldBeNil)
			So(deviceType, ShouldEqual, "ios")
			So(token, ShouldEqual, "devicetoken")
			So(userInfoID, ShouldEqual, "userid")
			So(lastRegisteredAt.Unix(), ShouldEqual, 1136214245)
		})

		Convey("updates an existing Device", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldBeNil)

			device.Token = "anotherdevicetoken"
			So(c.SaveDevice(&device), ShouldBeNil)

			var (
				deviceType, token, userInfoID string
				lastRegisteredAt              time.Time
			)
			err = c.QueryRowx("SELECT type, token, user_id, last_registered_at FROM app_com_oursky_skygear._device WHERE id = 'deviceid'").
				Scan(&deviceType, &token, &userInfoID, &lastRegisteredAt)
			So(err, ShouldBeNil)
			So(deviceType, ShouldEqual, "ios")
			So(token, ShouldEqual, "anotherdevicetoken")
			So(userInfoID, ShouldEqual, "userid")
			So(lastRegisteredAt.Unix(), ShouldEqual, 1136214245)
		})

		Convey("cannot save Device without id", func() {
			device := skydb.Device{
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("cannot save Device without type", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("can save Device without token", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "pubsub",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldBeNil)
		})

		Convey("cannot save Device without user id", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("cannot save Device without last_registered_at", func() {
			device := skydb.Device{
				ID:         "deviceid",
				Type:       "ios",
				Token:      "devicetoken",
				UserInfoID: "userid",
			}

			err := c.SaveDevice(&device)
			So(err, ShouldNotBeNil)
		})

		Convey("deletes an existing record", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			err := c.DeleteDevice("deviceid")
			So(err, ShouldBeNil)

			var count int
			err = c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._device WHERE id = 'deviceid'").Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("deletes an existing record by token", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			err := c.DeleteDeviceByToken("devicetoken", skydb.ZeroTime)
			So(err, ShouldBeNil)

			var count int
			err = c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._device WHERE id = 'deviceid'").Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("fails to delete an existing record with a later LastRegisteredAt", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			err := c.DeleteDeviceByToken("devicetoken", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(err, ShouldEqual, skydb.ErrDeviceNotFound)
		})

		Convey("deletes existing empty records", func() {
			device0 := skydb.Device{
				ID:               "deviceid0",
				Type:             "ios",
				Token:            "DEVICE_TOKEN",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			device1 := skydb.Device{
				ID:               "deviceid1",
				Type:             "ios",
				Token:            "",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			device2 := skydb.Device{
				ID:               "deviceid2",
				Type:             "ios",
				Token:            "",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device0), ShouldBeNil)
			So(c.SaveDevice(&device1), ShouldBeNil)
			So(c.SaveDevice(&device2), ShouldBeNil)

			err := c.DeleteEmptyDevicesByTime(skydb.ZeroTime)
			So(err, ShouldBeNil)

			var count int
			err = c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._device").Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})

		Convey("deletes existing empty records before a date", func() {
			device0 := skydb.Device{
				ID:               "deviceid0",
				Type:             "ios",
				Token:            "",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 4, 59, time.UTC),
			}
			device1 := skydb.Device{
				ID:               "deviceid1",
				Type:             "ios",
				Token:            "DEVICE_TOKEN",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device0), ShouldBeNil)
			So(c.SaveDevice(&device1), ShouldBeNil)

			err := c.DeleteEmptyDevicesByTime(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(err, ShouldBeNil)

			device := skydb.Device{}
			So(c.GetDevice("deviceid0", &device), ShouldEqual, skydb.ErrDeviceNotFound)
			So(c.GetDevice("deviceid1", &device), ShouldBeNil)
			So(device, ShouldResemble, device1)
		})

		Convey("fails to delete an existing record by type with a later LastRegisteredAt", func() {
			device := skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			err := c.DeleteEmptyDevicesByTime(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(err, ShouldEqual, skydb.ErrDeviceNotFound)
		})

		Convey("query devices by user", func() {
			device := skydb.Device{
				ID:               "device",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(c.SaveDevice(&device), ShouldBeNil)

			devices, err := c.QueryDevicesByUser("userid")
			So(err, ShouldBeNil)
			So(len(devices), ShouldEqual, 1)
			So(devices[0], ShouldResemble, skydb.Device{
				ID:               "device",
				Type:             "ios",
				Token:            "devicetoken",
				UserInfoID:       "userid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})

			devices, err = c.QueryDevicesByUser("nonexistent")
			So(err, ShouldBeNil)
			So(len(devices), ShouldEqual, 0)
		})
	})
}
