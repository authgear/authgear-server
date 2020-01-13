package apiversion

import (
	"fmt"
	"regexp"
	"strconv"
)

// MajorVersion is the current major API Version.
const MajorVersion = 2

// MinorVersion is the current minor API Version.
const MinorVersion = 0

// APIVersion is the current API Version.
var APIVersion = Format(MajorVersion, MinorVersion)

var regexpAPIVersion = regexp.MustCompile(`^v(\d+)\.(\d+)$`)

// Format formats major and minor into `v<major>.<minor>`.
func Format(major, minor int) string {
	return fmt.Sprintf("v%d.%d", major, minor)
}

// Parse parses API version into major and minor.
func Parse(apiVersion string) (major int, minor int, ok bool) {
	output := regexpAPIVersion.FindAllStringSubmatch(apiVersion, -1)
	if len(output) <= 0 {
		return
	}
	major, _ = strconv.Atoi(output[0][1])
	minor, _ = strconv.Atoi(output[0][2])
	ok = true
	return
}
