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
            "release": "22.0.0",
            "sysname": "Darwin",
            "version": "Darwin Kernel Version 22.0.0: Tue Aug 16 20:52:01 PDT 2022; root:xnu-8792.2.11.0.1~1/RELEASE_ARM64_T8101",
            "nodename": "rfc1123"
        },
        "NSBundle": {
            "CFBundleName": "Authgear Flutter",
            "CFBundleVersion": "1",
            "CFBundleExecutable": "Runner",
            "CFBundleIdentifier": "com.authgear.exampleapp.flutter",
            "CFBundleDisplayName": "Authgear Flutter",
            "CFBundleShortVersionString": "1.0.0"
        },
        "UIDevice": {
            "name": "iPhone",
            "model": "iPhone",
            "systemName": "iOS",
            "systemVersion": "16.0.2",
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
            "BOARD": "taro",
            "BRAND": "samsung",
            "MODEL": "SM-S9010",
            "DEVICE": "r0q",
            "DISPLAY": "SP1A.210812.016.S9010ZHU2AVF1",
            "PRODUCT": "r0qzhx",
            "VERSION": {
                "SDK": "31",
                "BASE_OS": "",
                "RELEASE": "12",
                "SDK_INT": "31",
                "CODENAME": "REL",
                "INCREMENTAL": "S9010ZHU2AVF1",
                "SECURITY_PATCH": "2022-06-01",
                "PREVIEW_SDK_INT": "0",
                "RELEASE_OR_CODENAME": "12"
            },
            "HARDWARE": "qcom",
            "MANUFACTURER": "samsung"
        },
        "Settings": {
            "Global": {
                "DEVICE_NAME": "Galaxy S22"
            },
            "Secure": {
                "ANDROID_ID": "08a54d3c65676f85",
                "bluetooth_name": "Galaxy S22"
            }
        },
        "PackageInfo": {
            "packageName": "com.authgear.exampleapp.flutter",
            "versionCode": "1",
            "versionName": "1.0",
            "longVersionCode": "1"
        },
        "ApplicationInfoLabel": "Authgear Flutter"
    }
}
`

func TestDevicePlatform(t *testing.T) {
	Convey("DevicePlatform", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := DevicePlatform(ios)
		So(actual, ShouldEqual, PlatformIOS)

		actual = DevicePlatform(android)
		So(actual, ShouldEqual, PlatformAndroid)

		So(DevicePlatform(nil), ShouldEqual, PlatformUnknown)
	})
}

func TestDeviceModelCodename(t *testing.T) {
	Convey("DeviceModelCodename", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := DeviceModelCodename(ios)
		So(actual, ShouldEqual, "iPhone13,1")

		actual = DeviceModelCodename(android)
		So(actual, ShouldEqual, "SM-S9010")

		So(DevicePlatform(nil), ShouldEqual, Platform(""))
	})
}

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
		So(actual, ShouldEqual, "samsung SM-S9010")

		So(DeviceModel(nil), ShouldEqual, "")

		// Allow unknown iPhone.
		So(DeviceModel(map[string]interface{}{
			"ios": map[string]interface{}{
				"uname": map[string]interface{}{
					"machine": "iPhone9999,9999",
				},
			},
		}), ShouldEqual, "iPhone9999,9999")

		// iPhone in 2023.
		So(DeviceModel(map[string]interface{}{
			"ios": map[string]interface{}{
				"uname": map[string]interface{}{
					"machine": "iPhone16,2",
				},
			},
		}), ShouldEqual, "iPhone 15 Pro Max")

		// iPhone in 2024.
		So(DeviceModel(map[string]interface{}{
			"ios": map[string]interface{}{
				"uname": map[string]interface{}{
					"machine": "iPhone17,2",
				},
			},
		}), ShouldEqual, "iPhone 16 Pro Max")
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
		So(actual, ShouldEqual, "Galaxy S22")

		So(DeviceName(nil), ShouldEqual, "")
	})
}

func TestApplicationName(t *testing.T) {
	Convey("ApplicationName", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := ApplicationName(ios)
		So(actual, ShouldEqual, "Authgear Flutter")

		actual = ApplicationName(android)
		So(actual, ShouldEqual, "Authgear Flutter")

		So(ApplicationName(nil), ShouldEqual, "")
	})
}

func TestApplicationID(t *testing.T) {
	Convey("ApplicationID", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual := ApplicationID(ios)
		So(actual, ShouldEqual, "com.authgear.exampleapp.flutter")

		actual = ApplicationID(android)
		So(actual, ShouldEqual, "com.authgear.exampleapp.flutter")

		So(ApplicationName(nil), ShouldEqual, "")
	})
}

func TestProbablySame(t *testing.T) {
	Convey("ProbablySame", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		So(ProbablySame(nil, nil), ShouldBeFalse)
		So(ProbablySame(ios, nil), ShouldBeFalse)
		So(ProbablySame(android, nil), ShouldBeFalse)
		So(ProbablySame(nil, ios), ShouldBeFalse)
		So(ProbablySame(nil, android), ShouldBeFalse)
		So(ProbablySame(ios, android), ShouldBeFalse)
		So(ProbablySame(ios, ios), ShouldBeTrue)
		So(ProbablySame(android, android), ShouldBeTrue)
	})
}
