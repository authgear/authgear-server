package web

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRecognizedMobileDevice(t *testing.T) {
	Convey("GetRecognizedMobileDevice", t, func() {
		Convey("should recognize iOS devices correctly", func() {
			iPadUserAgent := "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Mobile/15E148 Safari/604.1"
			iPhoneUserAgent := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1"

			device1, foundDevice1 := GetRecognizedMobileDevice(iPadUserAgent)
			device2, foundDevice2 := GetRecognizedMobileDevice(iPhoneUserAgent)
			So(device1, ShouldEqual, RecognizedMobileDeviceIOS)
			So(foundDevice1, ShouldBeTrue)
			So(device2, ShouldEqual, RecognizedMobileDeviceIOS)
			So(foundDevice2, ShouldBeTrue)
		})
		Convey("should recognize android device correctly", func() {
			androidUserAgent := "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Mobile Safari/537.36"

			device1, foundDevice1 := GetRecognizedMobileDevice(androidUserAgent)
			So(device1, ShouldEqual, RecognizedMobileDeviceChromeAndroid)
			So(foundDevice1, ShouldBeTrue)
		})
		Convey("should recognize samsung device correctly", func() {
			samsungUserAgent := "Mozilla/5.0 (Linux; Android 11; SAMSUNG SM-G973U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/14.2 Chrome/87.0.4280.141 Mobile Safari/537.36"

			device1, foundDevice1 := GetRecognizedMobileDevice(samsungUserAgent)
			So(device1, ShouldEqual, RecognizedMobileDeviceSamsung)
			So(foundDevice1, ShouldBeTrue)
		})
		Convey("should recognize chrome desktop device correctly", func() {
			chromeUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"

			device1, foundDevice1 := GetRecognizedMobileDevice(chromeUserAgent)
			So(device1, ShouldEqual, RecognizedMobileDeviceChrome)
			So(foundDevice1, ShouldBeTrue)
		})
		Convey("should return fallback case when cannot recognize device", func() {
			unknownPS5Device := "Mozilla/5.0 (PlayStation; PlayStation 5/2.26) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0 Safari/605.1.15"

			device1, foundDevice1 := GetRecognizedMobileDevice(unknownPS5Device)
			So(device1, ShouldBeEmpty)
			So(foundDevice1, ShouldBeFalse)
		})
		Convey("should not crash when user agent string is empty", func() {
			emptyString := ""

			device1, foundDevice1 := GetRecognizedMobileDevice(emptyString)
			So(device1, ShouldBeEmpty)
			So(foundDevice1, ShouldBeFalse)
		})
	})
}
