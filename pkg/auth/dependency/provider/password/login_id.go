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
	isValid(loginIDs []LoginID) bool
}

type defaultLoginIDChecker struct {
	loginIDsKeyWhitelist []string
}

func (c defaultLoginIDChecker) isValid(loginIDs []LoginID) bool {
	// default is empty list, allows any loginID keys
	allowAll := len(c.loginIDsKeyWhitelist) == 0
	for _, loginID := range loginIDs {
		// if loginIDsKeyWhitelist is not empty,
		// reject any loginIDKey that is not in the list
		if !allowAll &&
			!utils.StringSliceContains(c.loginIDsKeyWhitelist, loginID.Key) {
			return false
		}

		if loginID.Value == "" {
			return false
		}
	}
	return len(loginIDs) != 0
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
