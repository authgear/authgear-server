package loginid

import (
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Checker struct {
	Config             *config.LoginIDConfig
	TypeCheckerFactory *TypeCheckerFactory
}

func (c *Checker) Validate(specs []Spec) error {
	ctx := &validation.Context{}

	amounts := map[string]int{}
	for i, loginID := range specs {
		amounts[loginID.Key]++

		c.validateOne(ctx.Child(strconv.Itoa(i)), loginID)
	}

	for _, keyConfig := range c.Config.Keys {
		amount := amounts[keyConfig.Key]
		if amount > *keyConfig.MaxAmount {
			ctx.EmitErrorMessage("too many login IDs")
		}
	}

	if len(specs) == 0 {
		ctx.EmitError("required", nil)
	}

	return ctx.Error("invalid login IDs")
}

func (c *Checker) ValidateOne(loginID Spec) error {
	ctx := &validation.Context{}
	c.validateOne(ctx, loginID)
	return ctx.Error("invalid login ID")
}

func (c *Checker) validateOne(ctx *validation.Context, loginID Spec) {
	allowed := false
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginID.Key {
			if len(loginID.Value) > *keyConfig.MaxLength {
				ctx.EmitError("maxLength", map[string]interface{}{
					"expected": *keyConfig.MaxLength,
					"actual":   len(loginID.Value),
				})
				return
			}

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

	c.TypeCheckerFactory.NewChecker(loginID.Type).Validate(ctx, loginID.Value)
}

func (c *Checker) LoginIDKeyClaimName(loginIDKey string) (string, bool) {
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginIDKey {
			switch keyConfig.Type {
			case config.LoginIDKeyTypeEmail:
				return identity.StandardClaimEmail, true
			case config.LoginIDKeyTypePhone:
				return identity.StandardClaimPhoneNumber, true
			case config.LoginIDKeyTypeUsername:
				return identity.StandardClaimPreferredUsername, true
			default:
				return "", false
			}
		}
	}

	return "", false
}

func (c *Checker) CheckType(loginIDKey string, t config.LoginIDKeyType) bool {
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginIDKey {
			return keyConfig.Type == t
		}
	}
	return false
}
