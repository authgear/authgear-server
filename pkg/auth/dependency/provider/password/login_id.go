package password

import (
	"sort"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type LoginID struct {
	Key   string
	Value string
}

func ParseLoginIDs(rawLoginIDList []map[string]string) []LoginID {
	loginIDList := []LoginID{}
	for _, loginIDs := range rawLoginIDList {
		loginIDKeys := []string{}
		for key := range loginIDs {
			loginIDKeys = append(loginIDKeys, key)
		}
		sort.Strings(loginIDKeys)

		for _, loginIDKey := range loginIDKeys {
			loginIDList = append(loginIDList, LoginID{
				Key:   loginIDKey,
				Value: loginIDs[loginIDKey],
			})
		}
	}
	return loginIDList
}

type loginIDChecker interface {
	validate(loginIDs []LoginID) error
}

type defaultLoginIDChecker struct {
	loginIDsKeys map[string]config.LoginIDKeyConfiguration
}

func (c defaultLoginIDChecker) validate(loginIDs []LoginID) error {
	amounts := map[string]int{}
	for _, loginID := range loginIDs {
		_, allowed := c.loginIDsKeys[loginID.Key]
		if !allowed {
			return skyerr.NewInvalidArgument("login ID key is not allowed", []string{loginID.Key})
		}

		if loginID.Value == "" {
			return skyerr.NewInvalidArgument("login ID is empty", []string{loginID.Key})
		}
		amounts[loginID.Key]++
	}

	for key, keyConfig := range c.loginIDsKeys {
		amount := amounts[key]
		if amount > *keyConfig.Maximum || amount < *keyConfig.Minimum {
			return skyerr.NewInvalidArgument("login ID is not valid", []string{key})
		}
	}

	if len(loginIDs) == 0 {
		return skyerr.NewError(skyerr.InvalidArgument, "no login ID is present")
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
