package provider

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type PasswordCheckerProvider struct{}

func (p PasswordCheckerProvider) Provide(ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	return &audit.PasswordChecker{
		// TODO:
		// from tConfig
		PwMinLength: 6,
	}
}

type PasswordChecker interface {
	ValidatePassword(payload audit.ValidatePasswordPayload) skyerr.Error
}
