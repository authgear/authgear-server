package deviceinfo

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const ExampleIOS = `
{"ios": {"uname": {"machine": "iPhone13,1", "release": "20.3.0", "sysname": "Darwin", "version": "Darwin Kernel Version 20.3.0: Tue Jan  5 18:34:42 PST 2021; root:xnu-7195.80.35~2/RELEASE_ARM64_T8101", "nodename": "rfc1123"}, "UIDevice": {"name": "rfc1123", "model": "iPhone", "systemName": "iOS", "systemVersion": "14.4", "userInterfaceIdiom": "phone"}, "NSProcessInfo": {"isiOSAppOnMac": false, "isMacCatalystApp": false}}}
`

const ExampleAndroid = `
{"android": {"Build": {"BOARD": "blueline", "BRAND": "google", "MODEL": "Pixel 3", "DEVICE": "blueline", "DISPLAY": "RQ1A.201205.003", "PRODUCT": "blueline", "HARDWARE": "blueline", "MANUFACTURER": "Google"}}}
`

func TestFormat(t *testing.T) {
	Convey("Format", t, func() {
		var ios map[string]interface{}
		err := json.Unmarshal([]byte(ExampleIOS), &ios)
		So(err, ShouldBeNil)

		actual := Format(ios)
		So(actual, ShouldEqual, "iPhone 12 mini")

		var android map[string]interface{}
		err = json.Unmarshal([]byte(ExampleAndroid), &android)
		So(err, ShouldBeNil)

		actual = Format(android)
		So(actual, ShouldEqual, "Google Pixel 3")

		So(Format(nil), ShouldEqual, "")
	})
}
