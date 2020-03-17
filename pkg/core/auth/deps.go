package auth

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideAccessKeyMiddleware(c *config.TenantConfiguration) *AccessKeyMiddleware {
	m := &AccessKeyMiddleware{TenantConfig: c}
	return m
}

func ProvideAuthContextGetter(ctx context.Context) ContextGetter {
	return NewContextGetterWithContext(ctx)
}

var DependencySet = wire.NewSet(ProvideAccessKeyMiddleware, ProvideAuthContextGetter)
