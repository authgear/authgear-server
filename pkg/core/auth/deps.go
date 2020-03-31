package auth

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideAccessKeyMiddleware(c *config.TenantConfiguration) *AccessKeyMiddleware {
	m := &AccessKeyMiddleware{TenantConfig: c}
	return m
}

var DependencySet = wire.NewSet(ProvideAccessKeyMiddleware)
