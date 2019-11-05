package password

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/core/validation"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type LoginID struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (loginID LoginID) IsValid() bool {
	return len(loginID.Key) != 0 && len(loginID.Value) != 0
}

type loginIDChecker interface {
	validateOne(loginID LoginID) error
	validate(loginIDs []LoginID) error
	checkType(loginIDKey string, standardKey metadata.StandardKey) bool
	standardKey(loginIDKey string) (metadata.StandardKey, bool)
}

type defaultLoginIDChecker struct {
	loginIDsKeys []config.LoginIDKeyConfiguration
}

func (c defaultLoginIDChecker) validate(loginIDs []LoginID) error {
	amounts := map[string]int{}
	for i, loginID := range loginIDs {
		amounts[loginID.Key]++

		if err := c.validateOne(loginID); err != nil {
			if causes := validation.ErrorCauses(err); len(causes) > 0 {
				for j := range causes {
					causes[j].Pointer = fmt.Sprintf("/%d%s", i, causes[j].Pointer)
				}
				err = validation.NewValidationFailed("invalid login IDs", causes)
			}
			return err
		}
	}

	for _, keyConfig := range c.loginIDsKeys {
		amount := amounts[keyConfig.Key]
		if amount > *keyConfig.Maximum {
			return validation.NewValidationFailed("invalid login IDs", []validation.ErrorCause{{
				Kind:    validation.ErrorEntryAmount,
				Pointer: "",
				Message: "too many login IDs",
				Details: map[string]interface{}{"key": keyConfig.Key, "lte": *keyConfig.Maximum},
			}})
		} else if amount < *keyConfig.Minimum {
			return validation.NewValidationFailed("invalid login IDs", []validation.ErrorCause{{
				Kind:    validation.ErrorEntryAmount,
				Pointer: "",
				Message: "not enough login IDs",
				Details: map[string]interface{}{"key": keyConfig.Key, "gte": *keyConfig.Minimum},
			}})
		}
	}

	if len(loginIDs) == 0 {
		return validation.NewValidationFailed("invalid login IDs", []validation.ErrorCause{{
			Kind:    validation.ErrorRequired,
			Pointer: "",
			Message: "login ID is required",
		}})
	}

	return nil
}

func (c defaultLoginIDChecker) validateOne(loginID LoginID) error {
	allowed := false
	for _, keyConfig := range c.loginIDsKeys {
		if keyConfig.Key == loginID.Key {
			allowed = true
		}
	}
	if !allowed {
		return skyerr.NewInvalid("login ID key is not allowed")
	}

	if loginID.Value == "" {
		return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
			Kind:    validation.ErrorRequired,
			Pointer: "/value",
			Message: "login ID is required",
		}})
	}

	if err := c.validateLoginIDFormat(loginID); err != nil {
		return err
	}

	return nil
}

func (c defaultLoginIDChecker) standardKey(loginIDKey string) (key metadata.StandardKey, ok bool) {
	var config config.LoginIDKeyConfiguration
	for _, keyConfig := range c.loginIDsKeys {
		if keyConfig.Key == loginIDKey {
			config = keyConfig
			ok = true
			break
		}
	}
	if !ok {
		return
	}

	key, ok = config.Type.MetadataKey()
	return
}

func (c defaultLoginIDChecker) checkType(loginIDKey string, standardKey metadata.StandardKey) bool {
	loginIDKeyStandardKey, ok := c.standardKey(loginIDKey)
	return ok && loginIDKeyStandardKey == standardKey
}

func (c defaultLoginIDChecker) validateLoginIDFormat(id LoginID) error {
	key, ok := c.standardKey(id.Key)
	if !ok {
		return nil
	}

	ok = false
	var details map[string]interface{}
	switch key {
	case metadata.Email:
		ok = validation.Email{}.IsFormat(id.Value)
		details = map[string]interface{}{"format": "email"}
	case metadata.Phone:
		ok = validation.E164Phone{}.IsFormat(id.Value)
		details = map[string]interface{}{"format": "phone"}
	}
	if ok {
		return nil
	}
	return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
		Kind:    validation.ErrorStringFormat,
		Pointer: "/value",
		Message: "invalid login ID format",
		Details: details,
	}})
}

// this ensures that our structure conform to certain interfaces.
var (
	_ loginIDChecker = &defaultLoginIDChecker{}
)
