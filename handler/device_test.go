package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

type naiveConn struct {
	getid      string
	deleteid   string
	getdevice  *skydb.Device
	savedevice *skydb.Device
	geterr     error
	saveerr    error
	deleteerr  error
	skydb.Conn
}

func (conn *naiveConn) GetDevice(id string, device *skydb.Device) error {
	conn.getid = id
	if conn.geterr == nil {
		*device = *conn.getdevice
	}
	return conn.geterr
}

func (conn *naiveConn) SaveDevice(device *skydb.Device) error {
	conn.savedevice = device
	return conn.saveerr
}

func (conn *naiveConn) DeleteDevice(id string) error {
	conn.deleteid = id
	return conn.deleteerr
}

func TestDeviceRegisterHandler(t *testing.T) {
	Convey("DeviceRegisterHandler", t, func() {
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = timeNowUTC
		}()

		conn := naiveConn{}
		payload := router.Payload{
			DBConn:     &conn,
			UserInfoID: "userinfoid",
		}
		resp := router.Response{}

		Convey("creates new device", func() {
			payload.Data = map[string]interface{}{
				"type":         "ios",
				"device_token": "some-awesome-token",
			}

			DeviceRegisterHandler(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldNotBeEmpty)
			So(conn.savedevice, ShouldResemble, &skydb.Device{
				ID:               result.ID,
				Type:             "ios",
				Token:            "some-awesome-token",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("updates old device", func() {
			olddevice := skydb.Device{
				ID:               "deviceid",
				Type:             "android",
				Token:            "oldtoken",
				UserInfoID:       "olduserinfoid",
				LastRegisteredAt: time.Date(2005, 1, 2, 15, 4, 5, 0, time.UTC),
			}
			conn.getdevice = &olddevice

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
			}

			DeviceRegisterHandler(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldEqual, "deviceid")
			So(conn.getid, ShouldEqual, "deviceid")
			So(conn.savedevice, ShouldResemble, &skydb.Device{
				ID:               "deviceid",
				Type:             "ios",
				Token:            "newtoken",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("complains on empty device type", func() {
			payload.Data = map[string]interface{}{
				"device_token": "token",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewRequestInvalidErr(errors.New("empty device type")))
		})

		Convey("complains on invalid device type", func() {
			payload.Data = map[string]interface{}{
				"type":         "invalidtype",
				"device_token": "token",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewRequestInvalidErr(errors.New("unknown device type = invalidtype")))
		})

		Convey("does not complain on empty device token", func() {
			payload.Data = map[string]interface{}{
				"type": "android",
			}

			DeviceRegisterHandler(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldNotBeEmpty)
			So(conn.savedevice, ShouldResemble, &skydb.Device{
				ID:               result.ID,
				Type:             "android",
				Token:            "",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("complains on non-existed update", func() {
			conn.geterr = skydb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldEqual, skyerr.ErrDeviceNotFound)
		})

		Convey("complains on unknown device type", func() {
			conn.geterr = skydb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"type": "unknown-type",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(skyerr.Error)
			So(err, ShouldResemble, skyerr.NewRequestInvalidErr(errors.New("unknown device type = unknown-type")))
		})
	})
}
