package webapp

import (
	"regexp"
)

const WhatsappMessageOTPPrefix = "#"

var WhatsappMessageOTPRegex = regexp.MustCompile(`#(\d{6})`)

const WhatsappOTPPageQueryXDeviceTokenKey = "x_device_token"
