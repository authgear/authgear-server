package password

import "github.com/skygeario/skygear-server/pkg/core/utils"

type authDataChecker interface {
	isValid(authData map[string]string) bool
	isMatching(authData map[string]string) bool
}

type defaultAuthDataChecker struct {
	loginIDsKeyWhitelist []string
}

func (c defaultAuthDataChecker) isValid(authData map[string]string) bool {
	return len(toValidAuthDataMap(c.loginIDsKeyWhitelist, authData)) > 0
}

func (c defaultAuthDataChecker) isMatching(authData map[string]string) bool {
	return len(authData) == 1 && len(toValidAuthDataMap(c.loginIDsKeyWhitelist, authData)) == 1
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authDataChecker = &defaultAuthDataChecker{}
)

// toValidAuthDataMap converts authData to a list of authData depending on loginIDsKeyWhitelist
// example 1: loginIDsKeyWhitelist = []
// - allows to use any key
// - if authData is { "username": "john.doe" }, output is { "username": "john.doe" }
// - if authData is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { "username": "john.doe", "email": "john.doe@example.com" }
//
// example 2: loginIDsKeyWhitelist = ["username", "email"]
// - if authData is { "username": "john.doe" }, output is { "username": "john.doe" }
// - if authData is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { "username": "john.doe", "email": "john.doe@example.com" }
//
// example 3: loginIDsKeyWhitelist = ["email"]
// - if authData is { "username": "john.doe" }, output is {}
// - if authData is { "username": "john.doe", "email": "john.doe@example.com" }
//   output is { "email": "john.doe@example.com" }
//
func toValidAuthDataMap(loginIDsKeyWhitelist []string, authData map[string]string) map[string]string {
	outputs := make(map[string]string)
	for k, v := range authData {
		// default is empty list, allows any authData keys
		if (len(loginIDsKeyWhitelist) == 0 ||
			utils.StringSliceContains(loginIDsKeyWhitelist, k)) &&
			v != "" {
			outputs[k] = v
		}
	}
	return outputs
}
