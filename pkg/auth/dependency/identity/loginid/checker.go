package loginid

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type Checker struct {
	Config             *config.LoginIDConfig
	TypeCheckerFactory *TypeCheckerFactory
}

func (c *Checker) Validate(loginIDs []LoginID) error {
	amounts := map[string]int{}
	for i, loginID := range loginIDs {
		amounts[loginID.Key]++

		if err := c.ValidateOne(loginID); err != nil {
			if causes := validation.ErrorCauses(err); len(causes) > 0 {
				for j := range causes {
					causes[j].Pointer = fmt.Sprintf("/%d%s", i, causes[j].Pointer)
				}
				err = validation.NewValidationFailed("invalid login IDs", causes)
			}
			return err
		}
	}

	for _, keyConfig := range c.Config.Keys {
		amount := amounts[keyConfig.Key]
		if amount > *keyConfig.Maximum {
			return validation.NewValidationFailed("invalid login IDs", []validation.ErrorCause{{
				Kind:    validation.ErrorEntryAmount,
				Pointer: "",
				Message: "too many login IDs",
				Details: map[string]interface{}{"key": keyConfig.Key, "lte": *keyConfig.Maximum},
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

func (c *Checker) ValidateOne(loginID LoginID) error {
	allowed := false
	var loginIDType config.LoginIDKeyType
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginID.Key {
			loginIDType = keyConfig.Type
			allowed = true
		}
	}
	if !allowed {
		return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/key",
			Message: "login ID key is not allowed",
		}})
	}

	if loginID.Value == "" {
		return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
			Kind:    validation.ErrorRequired,
			Pointer: "/value",
			Message: "login ID is required",
		}})
	}

	if err := c.TypeCheckerFactory.NewChecker(loginIDType).Validate(loginID.Value); err != nil {
		return err
	}

	return nil
}

func (c *Checker) StandardKey(loginIDKey string) (key metadata.StandardKey, ok bool) {
	var config config.LoginIDKeyConfig
	for _, keyConfig := range c.Config.Keys {
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

func (c *Checker) CheckType(loginIDKey string, standardKey metadata.StandardKey) bool {
	loginIDKeyStandardKey, ok := c.StandardKey(loginIDKey)
	return ok && loginIDKeyStandardKey == standardKey
}
