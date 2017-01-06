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

package handler

import (
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

type naiveConn struct {
	devices                  map[string]skydb.Device
	mockGetError             error
	mockGetByTopicError      error
	mockSaveError            error
	mockDeleteError          error
	mockDeleteWithTokenError error
	skydb.Conn
}

func (conn *naiveConn) GetDevice(id string, device *skydb.Device) error {
	if conn.mockGetError != nil {
		return conn.mockGetError
	}

	targetDevice, ok := conn.devices[id]
	if !ok || targetDevice.Topic == "" {
		return skydb.ErrDeviceNotFound
	}

	*device = targetDevice
	return nil
}

func (conn *naiveConn) SaveDevice(device *skydb.Device) error {
	if conn.mockSaveError != nil {
		return conn.mockSaveError
	}

	deviceID := device.ID
	if deviceID != "" {
		conn.devices[deviceID] = *device
	}

	return nil
}

func (conn *naiveConn) DeleteDevice(id string) error {
	if conn.mockDeleteError != nil {
		return conn.mockDeleteError
	}

	delete(conn.devices, id)
	return nil
}

func (conn *naiveConn) DeleteDevicesByToken(token string, t time.Time) error {
	if conn.mockDeleteWithTokenError != nil {
		return conn.mockDeleteWithTokenError
	}

	newDevices := map[string]skydb.Device{}

	for perID, perDevice := range conn.devices {
		if perDevice.Token != token || (t != skydb.ZeroTime && perDevice.LastRegisteredAt.After(t)) {
			newDevices[perID] = perDevice
		}
	}

	conn.devices = newDevices

	return nil
}

func TestDeviceRegisterHandler(t *testing.T) {
	Convey("DeviceRegisterHandler", t, func() {
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = timeNowUTC
		}()

		conn := naiveConn{
			devices: map[string]skydb.Device{},
		}

		payload := router.Payload{
			DBConn:     &conn,
			UserInfoID: "userinfoid",
		}
		resp := router.Response{}

		Convey("creates new device", func() {
			payload.Data = map[string]interface{}{
				"type":         "ios",
				"device_token": "some-awesome-token",
				"topic":        "some-awesome-topic",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			resultID := result.ID
			So(resultID, ShouldNotBeEmpty)
			So(conn.devices[resultID], ShouldResemble, skydb.Device{
				ID:               resultID,
				Type:             "ios",
				Token:            "some-awesome-token",
				Topic:            "some-awesome-topic",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("updates old device", func() {
			olddevice := skydb.Device{
				ID:               "deviceid",
				Type:             "android",
				Token:            "oldtoken",
				Topic:            "oldtopic",
				UserInfoID:       "olduserinfoid",
				LastRegisteredAt: time.Date(2005, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(conn.SaveDevice(&olddevice), ShouldBeNil)

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
				"topic":        "newtopic",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			resultID := result.ID
			So(resultID, ShouldEqual, "deviceid")
			So(conn.devices[resultID], ShouldResemble, skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "newtoken",
				Topic:            "newtopic",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("remove devices with the same token when register", func() {
			existingDevice := skydb.Device{
				ID:               "existing_id",
				Type:             "ios",
				Token:            "existing_token",
				Topic:            "existing_topic",
				UserInfoID:       "existing_user",
				LastRegisteredAt: time.Date(2005, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			So(conn.SaveDevice(&existingDevice), ShouldBeNil)

			payload.Data = map[string]interface{}{
				"type":         "ios",
				"device_token": "existing_token",
				"topic":        "existing_topic",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			resultID := result.ID
			So(resultID, ShouldNotBeEmpty)
			So(conn.devices[resultID], ShouldResemble, skydb.Device{
				ID:               resultID,
				Type:             "ios",
				Token:            "existing_token",
				Topic:            "existing_topic",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
			So(conn.devices["existing_id"], ShouldResemble, skydb.Device{})
		})

		Convey("complains on empty device type", func() {
			payload.Data = map[string]interface{}{
				"device_token": "token",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewInvalidArgument("empty device type", []string{"type"}))
		})

		Convey("complains on invalid device type", func() {
			payload.Data = map[string]interface{}{
				"type":         "invalidtype",
				"device_token": "token",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewInvalidArgument("unknown device type = invalidtype", []string{"type"}))
		})

		Convey("does not complain on empty device token", func() {
			payload.Data = map[string]interface{}{
				"type":  "android",
				"topic": "some-topic",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			resultID := result.ID
			So(resultID, ShouldNotBeEmpty)
			So(conn.devices[resultID], ShouldResemble, skydb.Device{
				ID:               result.ID,
				Type:             "android",
				Token:            "",
				Topic:            "some-topic",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("complains on non-existed update", func() {
			conn.mockGetError = skydb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
				"topic":        "some-topic",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewError(
				skyerr.ResourceNotFound,
				"Device not found",
			))
		})

		Convey("complains on unknown device type", func() {
			conn.mockGetError = skydb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"type": "unknown-type",
			}

			handler := &DeviceRegisterHandler{}
			handler.Handle(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewInvalidArgument("unknown device type = unknown-type", []string{"type"}))
		})
	})
}

func TestDeviceUnregisterHandler(t *testing.T) {
	Convey("DeviceUnregisterHandler", t, func() {
		conn := naiveConn{
			devices: map[string]skydb.Device{
				"device_1": skydb.Device{
					ID:               "device_1",
					Type:             "ios",
					Token:            "device_token_1",
					Topic:            "device_topic_1",
					UserInfoID:       "user_id_1",
					LastRegisteredAt: time.Date(2016, 12, 16, 6, 54, 0, 0, time.UTC),
				},
				"device_2_1": skydb.Device{
					ID:               "device_2_1",
					Type:             "ios",
					Token:            "device_token_2",
					Topic:            "device_topic_2",
					UserInfoID:       "user_id_2",
					LastRegisteredAt: time.Date(2016, 12, 16, 6, 55, 0, 0, time.UTC),
				},
				"device_2_2": skydb.Device{
					ID:               "device_2_2",
					Type:             "ios",
					Token:            "device_token_2",
					Topic:            "device_topic_3",
					UserInfoID:       "user_id_3",
					LastRegisteredAt: time.Date(2016, 12, 16, 6, 56, 0, 0, time.UTC),
				},
			},
		}

		Convey("removes user id of target device", func() {
			payload := router.Payload{
				DBConn:     &conn,
				UserInfoID: "user_id_1",
				Data: map[string]interface{}{
					"id": "device_1",
				},
			}

			resp := router.Response{}
			handler := &DeviceUnregisterHandler{}

			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldEqual, "device_1")

			device := conn.devices["device_1"]
			So(device.ID, ShouldEqual, "device_1")
			So(device.Type, ShouldEqual, "ios")
			So(device.Token, ShouldEqual, "device_token_1")
			So(device.UserInfoID, ShouldBeEmpty)
			So(
				device.LastRegisteredAt,
				ShouldResemble,
				time.Date(2016, 12, 16, 6, 54, 0, 0, time.UTC),
			)
		})

		Convey("deletes other devices with the same token", func() {
			payload := router.Payload{
				DBConn:     &conn,
				UserInfoID: "user_id_2",
				Data: map[string]interface{}{
					"id": "device_2_1",
				},
			}

			resp := router.Response{}
			handler := &DeviceUnregisterHandler{}

			handler.Handle(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldEqual, "device_2_1")

			device := conn.devices["device_2_1"]
			So(device.ID, ShouldEqual, "device_2_1")
			So(device.Type, ShouldEqual, "ios")
			So(device.Token, ShouldEqual, "device_token_2")
			So(device.UserInfoID, ShouldBeEmpty)
			So(
				device.LastRegisteredAt,
				ShouldResemble,
				time.Date(2016, 12, 16, 6, 55, 0, 0, time.UTC),
			)

			_, ok := conn.devices["device_2_2"]
			So(ok, ShouldBeFalse)
		})

		Convey("complains on non-existed update", func() {
			payload := router.Payload{
				DBConn:     &conn,
				UserInfoID: "user_id_3",
				Data: map[string]interface{}{
					"id": "device_3",
				},
			}

			resp := router.Response{}
			handler := &DeviceUnregisterHandler{}

			handler.Handle(&payload, &resp)

			err := resp.Err
			So(err, ShouldResemble, skyerr.NewError(
				skyerr.ResourceNotFound,
				"Device not found",
			))
		})

		Convey("complains on empty device id", func() {
			payload := router.Payload{
				DBConn:     &conn,
				UserInfoID: "user_id_3",
				Data:       map[string]interface{}{},
			}

			resp := router.Response{}
			handler := &DeviceUnregisterHandler{}

			handler.Handle(&payload, &resp)

			err := resp.Err
			So(err, ShouldResemble, skyerr.NewInvalidArgument(
				"Missing device id",
				[]string{"id"},
			))
		})
	})
}
