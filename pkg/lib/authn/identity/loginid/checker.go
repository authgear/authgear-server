package loginid

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type CheckerOptions struct {
	EmailByPassBlocklistAllowlist bool
}

type Checker struct {
	Config             *config.LoginIDConfig
	TypeCheckerFactory *TypeCheckerFactory
}

func (c *Checker) ValidateOne(ctx context.Context, loginID identity.LoginIDSpec, options CheckerOptions) error {
	validationCtx := &validation.Context{}
	c.validateOne(ctx, validationCtx, loginID, options)
	return validationCtx.Error("invalid login ID")
}

func (c *Checker) validateOne(ctx context.Context, validationCtx *validation.Context, loginID identity.LoginIDSpec, options CheckerOptions) {
	originCtx := validationCtx
	validationCtx = validationCtx.Child("login_id")

	allowed := false
	for _, keyConfig := range c.Config.Keys {
		if keyConfig.Key == loginID.Key {
			if len(loginID.Value.TrimSpace()) > *keyConfig.MaxLength {
				validationCtx.EmitError("maxLength", map[string]interface{}{
					"expected": *keyConfig.MaxLength,
					"actual":   len(loginID.Value.TrimSpace()),
				})
				return
			}

			allowed = true
		}
	}
	if !allowed {
		validationCtx.EmitErrorMessage("login ID key is not allowed")
		return
	}

	if loginID.Value.TrimSpace() == "" {
		validationCtx.EmitError("required", nil)
		return
	}

	c.TypeCheckerFactory.NewChecker(loginID.Type, options).Validate(ctx, originCtx, loginID.Value.TrimSpace())
}
