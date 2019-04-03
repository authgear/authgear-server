package password

import "github.com/skygeario/skygear-server/pkg/core/utils"

type loginIDChecker interface {
	isValid(loginID map[string]string) bool
}

type defaultLoginIDChecker struct {
	loginIDsKeyWhitelist []string
}

func (c defaultLoginIDChecker) isValid(loginIDs map[string]string) bool {
	// default is empty list, allows any loginID keys
	allowAll := len(c.loginIDsKeyWhitelist) == 0
	for k, v := range loginIDs {
		// if loginIDsKeyWhitelist is not empty,
		// reject any loginIDKey that is not in the list
		if !allowAll &&
			!utils.StringSliceContains(c.loginIDsKeyWhitelist, k) {
			return false
		}

		if v == "" {
			return false
		}
	}
	return len(loginIDs) != 0
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
