package handler

import (
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	. "github.com/oursky/skygear/ourtest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
)

func TestPushToDevice(t *testing.T) {
	Convey("push to device", t, func() {
		testdevice := skydb.Device{
			ID:         "device",
			Type:       "ios",
			Token:      "token",
			UserInfoID: "userid",
		}
		conn := simpleDeviceConn{
			devices: []skydb.Device{testdevice},
		}

		r := handlertest.NewSingleRouteRouter(PushToDeviceHandler, func(p *router.Payload) {
			p.DBConn = &conn
		})

		originalSendFunc := sendPushNotification
		defer func() {
			sendPushNotification = originalSendFunc
		}()

		Convey("push to single device", func() {
			called := false
			sendPushNotification = func(sender push.Sender, device *skydb.Device, m push.Mapper) {
				So(device, ShouldResemble, &testdevice)
				So(m.Map(), ShouldResemble, map[string]interface{}{
					"aps": map[string]interface{}{
						"alert": "This is a message.",
						"sound": "sosumi.mp3",
					},
					"acme": "interesting",
				})
				called = true
			}
			resp := r.POST(`{
					"device_ids": ["device"],
					"notification": {
						"aps": {
							"alert": "This is a message.",
							"sound": "sosumi.mp3"
						},
						"acme": "interesting"
					}
				}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "device"
	}]
}`)
			So(called, ShouldBeTrue)
		})

		Convey("push to non-existent device", func() {
			called := false
			sendPushNotification = func(sender push.Sender, device *skydb.Device, m push.Mapper) {
				called = true
			}
			resp := r.POST(`{
						"device_ids": ["nonexistent"],
						"notification": {
							"aps": {
								"alert": "This is a message.",
								"sound": "sosumi.mp3"
							},
							"acme": "interesting"
						}
					}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "nonexistent",
		"_type": "error",
		"message": "cannot find device \"nonexistent\"",
		"type": "ResourceNotFound",
		"code": 101,
		"info": {"id": "nonexistent"}
	}]
}`)
			So(called, ShouldBeFalse)
		})
	})

}

func TestPushToUser(t *testing.T) {
	Convey("push to user", t, func() {
		testdevice1 := skydb.Device{
			ID:         "device1",
			Type:       "ios",
			Token:      "token",
			UserInfoID: "johndoe",
		}
		testdevice2 := skydb.Device{
			ID:         "device2",
			Type:       "ios",
			Token:      "token",
			UserInfoID: "johndoe",
		}
		testdevice3 := skydb.Device{
			ID:         "device2",
			Type:       "ios",
			Token:      "token",
			UserInfoID: "janedoe",
		}
		conn := simpleDeviceConn{
			devices: []skydb.Device{testdevice1, testdevice2, testdevice3},
		}

		r := handlertest.NewSingleRouteRouter(PushToUserHandler, func(p *router.Payload) {
			p.DBConn = &conn
		})

		originalSendFunc := sendPushNotification
		defer func() {
			sendPushNotification = originalSendFunc
		}()

		Convey("push to single user", func() {
			sentDevices := make([]skydb.Device, 0)
			sendPushNotification = func(sender push.Sender, device *skydb.Device, m push.Mapper) {
				So(m.Map(), ShouldResemble, map[string]interface{}{
					"aps": map[string]interface{}{
						"alert": "This is a message.",
						"sound": "sosumi.mp3",
					},
					"acme": "interesting",
				})
				sentDevices = append(sentDevices, *device)
			}
			resp := r.POST(`{
					"user_ids": ["johndoe"],
					"notification": {
						"aps": {
							"alert": "This is a message.",
							"sound": "sosumi.mp3"
						},
						"acme": "interesting"
					}
				}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{"_id":"johndoe"}]
}`)

			So(len(sentDevices), ShouldEqual, 2)
			So(sentDevices[0], ShouldResemble, testdevice1)
			So(sentDevices[1], ShouldResemble, testdevice2)
		})

		Convey("push to non-existent user", func() {
			called := false
			sendPushNotification = func(sender push.Sender, device *skydb.Device, m push.Mapper) {
				called = true
			}
			resp := r.POST(`{
					"user_ids": ["nonexistent"],
					"notification": {
						"aps": {
							"alert": "This is a message.",
							"sound": "sosumi.mp3"
						},
						"acme": "interesting"
					}
				}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "nonexistent",
		"_type": "error",
		"message": "cannot find user \"nonexistent\"",
		"type": "ResourceNotFound",
		"code": 101,
		"info": {"id": "nonexistent"}
	}]
}`)
			So(called, ShouldBeFalse)
		})
	})

}

type simpleDeviceConn struct {
	devices []skydb.Device
	skydb.Conn
}

func (conn *simpleDeviceConn) GetDevice(id string, device *skydb.Device) error {
	for _, prospectiveDevice := range conn.devices {
		if prospectiveDevice.ID == id {
			*device = prospectiveDevice
			return nil
		}
	}
	return skydb.ErrDeviceNotFound
}

func (conn *simpleDeviceConn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	result := make([]skydb.Device, 0)
	for _, prospectiveDevice := range conn.devices {
		if prospectiveDevice.UserInfoID == user {
			result = append(result, prospectiveDevice)
		}
	}
	if len(result) == 0 {
		return nil, skydb.ErrUserNotFound
	} else {
		return result, nil
	}
}
