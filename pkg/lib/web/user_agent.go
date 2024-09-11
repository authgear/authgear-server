package web

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// The name is borrowed from https://github.com/browserslist/browserslist
type RecognizedMobileDeviceType string

type RecognizedMobileDevice struct {
	Type                 RecognizedMobileDeviceType
	OSVersionMajorString string
	OSVersionMinorString string
	OSVersionPatchString string
	OSVersionMajorInt    int
	OSVersionMinorInt    int
	OSVersionPatchInt    int
}

const (
	RecognizedMobileDeviceTypeIOS           = "iOS"
	RecognizedMobileDeviceTypeChromeAndroid = "ChromeAndroid"
	RecognizedMobileDeviceTypeChrome        = "Chrome"
	RecognizedMobileDeviceTypeSamsung       = "Samsung"
)

func GetRecognizedMobileDevice(ua string) (device RecognizedMobileDevice) {
	client := model.ParseUserAgentRaw(ua)

	device.OSVersionMajorString = client.Os.Major
	device.OSVersionMinorString = client.Os.Minor
	device.OSVersionPatchString = client.Os.Patch
	device.OSVersionMajorInt, _ = strconv.Atoi(device.OSVersionMajorString)
	device.OSVersionMinorInt, _ = strconv.Atoi(device.OSVersionMinorString)
	device.OSVersionPatchInt, _ = strconv.Atoi(device.OSVersionPatchString)

	if client.Os.Family == "iOS" {
		device.Type = RecognizedMobileDeviceTypeIOS
	}
	if client.Os.Family == "Android" {
		device.Type = RecognizedMobileDeviceTypeChromeAndroid
	}
	if client.Os.Family == "Android" && client.UserAgent.Family == "Samsung Internet" {
		device.Type = RecognizedMobileDeviceTypeSamsung
	}
	if client.UserAgent.Family == "Chrome" {
		device.Type = RecognizedMobileDeviceTypeChrome
	}

	return device
}
