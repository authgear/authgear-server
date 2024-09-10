package web

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

// The name is borrowed from https://github.com/browserslist/browserslist
type RecognizedMobileDeviceType string

const (
	RecognizedMobileDeviceIOS           = "iOS"
	RecognizedMobileDeviceChromeAndroid = "ChromeAndroid"
	RecognizedMobileDeviceChrome        = "Chrome"
	RecognizedMobileDeviceSamsung       = "Samsung"
)

func GetRecognizedMobileDevice(ua string) (string, bool) {
	client := model.ParseUserAgentRaw(ua)
	if client.Os.Family == "iOS" {
		return RecognizedMobileDeviceIOS, true
	}
	if client.Os.Family == "Android" && client.Device.Brand == "Samsung" {
		return RecognizedMobileDeviceSamsung, true
	}
	if client.Os.Family == "Android" {
		return RecognizedMobileDeviceChromeAndroid, true
	}
	if client.UserAgent.Family == "Chrome" {
		return RecognizedMobileDeviceChrome, true
	}

	return "", false
}
