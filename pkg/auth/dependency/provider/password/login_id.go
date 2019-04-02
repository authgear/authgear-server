package password

import "github.com/skygeario/skygear-server/pkg/core/utils"

type loginIDChecker interface {
	isValid(loginID map[string]string) bool
	isMatching(loginID map[string]string) bool
}

type defaultLoginIDChecker struct {
	loginIDsKeyWhitelist []string
}

func (c defaultLoginIDChecker) isValid(loginID map[string]string) bool {
	return len(toValidLoginIDMap(c.loginIDsKeyWhitelist, loginID)) > 0
}

func (c defaultLoginIDChecker) isMatching(loginID map[string]string) bool {
	return len(loginID) == 1 && len(toValidLoginIDMap(c.loginIDsKeyWhitelist, loginID)) == 1
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)

// toValidLoginIDMap converts loginID to a list of loginID depending on loginIDsKeyWhitelist
// example 1: loginIDsKeyWhitelist = []
// - allows to use any key
// - if loginID is { "username": "john.doe" }, output is { "username": "john.doe" }
// - if loginID is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { "username": "john.doe", "email": "john.doe@example.com" }
//
// example 2: loginIDsKeyWhitelist = ["username", "email"]
// - if loginID is { "username": "john.doe" }, output is { "username": "john.doe" }
// - if loginID is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { "username": "john.doe", "email": "john.doe@example.com" }
//
// example 3: loginIDsKeyWhitelist = ["email"]
// - if loginID is { "username": "john.doe" }, output is {}
// - if loginID is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { } // username is not in the list
//
func toValidLoginIDMap(loginIDsKeyWhitelist []string, loginID map[string]string) map[string]string {
	for k, v := range loginID {
		// default is empty list, allows any loginID keys
		// if loginIDsKeyWhitelist is not empty, reject any loginIDKey that is not in the list
		if len(loginIDsKeyWhitelist) != 0 &&
			!utils.StringSliceContains(loginIDsKeyWhitelist, k) {
			return map[string]string{}
		}

		if v == "" {
			return map[string]string{}
		}
	}
	return loginID
}
