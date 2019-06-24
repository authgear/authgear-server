package name

import (
	"errors"
	"regexp"
)

var (
	// ErrInvalidAppName tells the accepted format of app name.
	// App name will be used in domain name so it must
	// start with a letter.
	// Kubernetes only accepts name of 63 letter long,
	// so we have to limit the length.
	ErrInvalidAppName = errors.New("App name must be ^[a-zA-Z][a-zA-Z0-9]{0,11}$")
)

var (
	regexAppName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,11}$`)
)

func ValidateAppName(appName string) error {
	if !regexAppName.MatchString(appName) {
		return ErrInvalidAppName
	}
	return nil
}
