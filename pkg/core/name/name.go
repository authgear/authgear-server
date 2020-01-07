package name

import (
	"regexp"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// A DNS label must have at most 63 characters
// We limit the maximum length to 40, reserving some space for internal use,
// e.g. k8s pod suffix
const appNameFormat = `^[a-z0-9]([-a-z0-9]{0,38}[a-z0-9])?$`

var (
	// ErrInvalidAppName tells the accepted format of app name.
	// App name will be used in domain name so it must be a valid DNS label.
	ErrInvalidAppName = errors.New("must contain only lowercase alphanumeric characters/dash, at most 40 characters, and not start/end with dash")
)
var (
	appNameRegex = regexp.MustCompile(appNameFormat)
)

func ValidateAppName(appName string) error {
	if !appNameRegex.MatchString(appName) {
		return ErrInvalidAppName
	}
	return nil
}
