package password

import "github.com/skygeario/skygear-server/pkg/core/utils"

type LoginID struct {
	Key   string
	Value string
}

func ParseLoginIDs(rawLoginIDList []map[string]string) []LoginID {
	loginIDList := []LoginID{}
	for _, loginIDs := range rawLoginIDList {
		for loginIDKey, loginID := range loginIDs {
			loginIDList = append(loginIDList, LoginID{
				Key:   loginIDKey,
				Value: loginID,
			})
		}
	}
	return loginIDList
}

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
