package handler

import (
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
		conn := naiveConn{}
		payload := router.Payload{
			DBConn: &conn,
			UserInfo: &oddb.UserInfo{
				ID: "userinfoid",
			},
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
				ID:         result.ID,
				Type:       "ios",
				Token:      "some-awesome-token",
				UserInfoID: "userinfoid",
			})
		})

		Convey("updates old device", func() {
			olddevice := oddb.Device{
				ID:         "deviceid",
				Type:       "android",
				Token:      "oldtoken",
				UserInfoID: "olduserinfoid",
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
				ID:         "deviceid",
				Type:       "ios",
				Token:      "newtoken",
				UserInfoID: "userinfoid",
			})
		})

		Convey("complains on empty device type", func() {
			payload.Data = map[string]interface{}{
				"device_token": "token",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Result.(oderr.Error)
			So(err.Code(), ShouldEqual, oderr.RequestInvalidErr)
		})

		Convey("complains on invalid device type", func() {
			payload.Data = map[string]interface{}{
				"type":         "invalidtype",
				"device_token": "token",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Result.(oderr.Error)
			So(err.Code(), ShouldEqual, oderr.RequestInvalidErr)
		})

		Convey("complains on empty device token", func() {
			payload.Data = map[string]interface{}{
				"type": "android",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Result.(oderr.Error)
			So(err.Code(), ShouldEqual, oderr.RequestInvalidErr)
		})

		Convey("complains on non-existed update", func() {
			conn.geterr = oddb.ErrDeviceNotFound

			payload.Data = map[string]interface{}{
				"id":           "deviceid",
				"type":         "ios",
				"device_token": "newtoken",
			}

			DeviceRegisterHandler(&payload, &resp)

			err := resp.Result.(oderr.Error)
			So(err.Code(), ShouldEqual, oderr.RequestInvalidErr)
		})
	})
}
