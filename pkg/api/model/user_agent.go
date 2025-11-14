package model

import (
	"fmt"

	"github.com/ua-parser/uap-go/uaparser"
)

var uaParser = uaparser.NewFromSaved()

type UserAgent struct {
	Raw         string `json:"raw"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	OS          string `json:"os"`
	OSVersion   string `json:"os_version"`
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
}

func (u *UserAgent) Format() string {
	var out string
	if u.Name != "" {
		out += u.Name
	}
	if u.Version != "" {
		out += " " + u.Version
	}
	return out
}

func ParseUserAgent(ua string) (mUA UserAgent) {
	mUA.Raw = ua

	client := uaParser.Parse(ua)
	if client.UserAgent.Family != "Other" {
		mUA.Name = client.UserAgent.Family
		mUA.Version = client.UserAgent.ToVersionString()
	}
	if client.Device.Family != "Other" {
		mUA.DeviceModel = fmt.Sprintf("%s %s", client.Device.Brand, client.Device.Model)
	}
	if client.Os.Family != "Other" {
		mUA.OS = client.Os.Family
		mUA.OSVersion = client.Os.ToVersionString()
	}

	return mUA
}

// The name is borrowed from https://github.com/browserslist/browserslist
type RecognizedMobileDevice string

const (
	RecognizedMobileDeviceIOS           = "iOS"
	RecognizedMobileDeviceChromeAndroid = "ChromeAndroid"
	RecognizedMobileDeviceChrome        = "Chrome"
	RecognizedMobileDeviceSamsung       = "Samsung"
)

func GetRecognizedMobileDevice(ua string) (string, bool) {
	client := uaParser.Parse(ua)
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
