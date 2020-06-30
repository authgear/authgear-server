package loginid

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/auth/metadata"
	"github.com/authgear/authgear-server/pkg/validation"
)

type Checker struct {
	Config             *config.LoginIDConfig
	TypeCheckerFactory *TypeCheckerFactory
}

func (c *Checker) Validate(loginIDs []LoginID) error {
	ctx := &validation.Context{}

	amounts := map[string]int{}
	for i, loginID := range loginIDs {
		amounts[loginID.Key]++

		c.validateOne(ctx.Child(strconv.Itoa(i)), loginID)
	}

	for _, keyConfig := range c.Config.Keys {
		amount := amounts[keyConfig.Key]
		if amount > *keyConfig.Maximum {
			ctx.EmitErrorMessage("too many login IDs")
		}
	}

	if len(loginIDs) == 0 {
		ctx.EmitError("required", nil)
	}

	return ctx.Error("invalid login IDs")
}

func (c *Checker) ValidateOne(loginID LoginID) error {
	ctx := &validation.Context{}
	c.validateOne(ctx, loginID)
	return ctx.Error("invalid login ID")
}

func (c *Checker) validateOne(ctx *validation.Context, loginID LoginID) {
	allowed := false
	var loginIDType config.LoginIDKeyType
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginID.Key {
			loginIDType = keyConfig.Type
			allowed = true
		}
	}
	if !allowed {
		ctx.EmitErrorMessage("login ID key is not allowed")
		return
	}

	if loginID.Value == "" {
		ctx.EmitError("required", nil)
		return
	}

	c.TypeCheckerFactory.NewChecker(loginIDType).Validate(ctx, loginID.Value)
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
