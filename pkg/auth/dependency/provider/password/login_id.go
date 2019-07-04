package password

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type LoginID struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (loginID LoginID) IsValid() bool {
	return len(loginID.Key) != 0 && len(loginID.Value) != 0
}

type loginIDChecker interface {
	validate(loginIDs []LoginID) error
	checkType(loginIDKey string, standardKey metadata.StandardKey) bool
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

func (c defaultLoginIDChecker) checkType(loginIDKey string, standardKey metadata.StandardKey) bool {
	config, ok := c.loginIDsKeys[loginIDKey]
	if !ok {
		return false
	}

	configStandardKey, ok := config.Type.MetadataKey()
	return ok && configStandardKey == standardKey
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
