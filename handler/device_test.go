package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
)

type naiveConn struct {
	getid      string
	deleteid   string
	getdevice  *oddb.Device
	savedevice *oddb.Device
	geterr     error
	saveerr    error
	deleteerr  error
	oddb.Conn
}

func (conn *naiveConn) GetDevice(id string, device *oddb.Device) error {
	conn.getid = id
	if conn.geterr == nil {
		*device = *conn.getdevice
	}
	return conn.geterr
}

func (conn *naiveConn) SaveDevice(device *oddb.Device) error {
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
			So(conn.savedevice, ShouldResemble, &oddb.Device{
				ID:               result.ID,
				Type:             "ios",
				Token:            "some-awesome-token",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("updates old device", func() {
			olddevice := oddb.Device{
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
			So(conn.savedevice, ShouldResemble, &oddb.Device{
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

			err := resp.Err.(oderr.Error)
			So(err, ShouldResemble, oderr.NewRequestInvalidErr(errors.New("empty device type")))
		})

		Convey("complains on invalid device type", func() {
			payload.Data = map[string]interface{}{
				"type":         "invalidtype",
				"device_token": "token",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(oderr.Error)
			So(err, ShouldResemble, oderr.NewRequestInvalidErr(errors.New("unknown device type = invalidtype")))
		})

		Convey("does not complain on empty device token", func() {
			payload.Data = map[string]interface{}{
				"type": "android",
			}

			DeviceRegisterHandler(&payload, &resp)

			result := resp.Result.(DeviceReigsterResult)
			So(result.ID, ShouldNotBeEmpty)
			So(conn.savedevice, ShouldResemble, &oddb.Device{
				ID:               result.ID,
				Type:             "android",
				Token:            "",
				UserInfoID:       "userinfoid",
				LastRegisteredAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			})
		})

		Convey("complains on non-existed update", func() {
			conn.geterr = oddb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(oderr.Error)
			So(err, ShouldEqual, oderr.ErrDeviceNotFound)
		})

		Convey("complains on unknown device type", func() {
			conn.geterr = oddb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"type": "unknown-type",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Err.(oderr.Error)
			So(err, ShouldResemble, oderr.NewRequestInvalidErr(errors.New("unknown device type = unknown-type")))
		})
	})
}
