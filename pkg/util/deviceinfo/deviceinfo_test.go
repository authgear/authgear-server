package deviceinfo

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const ExampleIOS = `
{
	"ios": {
		"uname": {
			"machine": "iPhone13,1",
			"release": "20.3.0",
			"sysname": "Darwin",
			"version": "Darwin Kernel Version 20.3.0: Tue Jan  5 18:34:42 PST 2021; root:xnu-7195.80.35~2/RELEASE_ARM64_T8101",
			"nodename": "rfc1123"
		},
		"UIDevice": {
			"name": "rfc1123",
			"model": "iPhone",
			"systemName": "iOS",
			"systemVersion": "14.4",
			"userInterfaceIdiom": "phone"
		},
		"NSProcessInfo": {
			"isiOSAppOnMac": false,
			"isMacCatalystApp": false
		}
	}
}
`

const ExampleAndroid = `
{
	"android": {
		"Build": {
			"BOARD": "blueline",
			"BRAND": "google",
			"MODEL": "Pixel 3",
			"DEVICE": "blueline",
			"DISPLAY": "RQ1A.201205.003",
			"PRODUCT": "blueline",
			"HARDWARE": "blueline",
			"MANUFACTURER": "Google"
		},
		"Settings": {
			"Global": {
				"DEVICE_NAME": "myandroid"
			},
			"Secure": {
				"bluetooth_name": "myandroid"
			}
		}
	}
}
`

func TestDeviceModel(t *testing.T) {
	Convey("DeviceModel", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := DeviceModel(ios)
		So(actual, ShouldEqual, "iPhone 12 mini")

		actual = DeviceModel(android)
		So(actual, ShouldEqual, "Google Pixel 3")

		So(DeviceModel(nil), ShouldEqual, "")
	})
}

func TestDeviceName(t *testing.T) {
	Convey("DeviceName", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := DeviceName(ios)
		So(actual, ShouldEqual, "rfc1123")

		actual = DeviceName(android)
		So(actual, ShouldEqual, "myandroid")

		So(DeviceName(nil), ShouldEqual, "")
	})
}
