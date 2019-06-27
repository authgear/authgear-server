package password

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

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
	loginIDsKeys map[string]config.LoginIDKeyConfiguration
}

func (c defaultLoginIDChecker) isValid(loginIDs []LoginID) bool {
	amounts := map[string]int{}
	for _, loginID := range loginIDs {
		_, allowed := c.loginIDsKeys[loginID.Key]
		if !allowed {
			return false
		}

		if loginID.Value == "" {
			return false
		}
		amounts[loginID.Key]++
	}

	for key, keyConfig := range c.loginIDsKeys {
		amount := amounts[key]
		if amount > *keyConfig.Maximum || amount < *keyConfig.Minimum {
			return false
		}
	}

	return len(loginIDs) != 0
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
