package auth

import (
	"context"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideAccessKeyMiddleware(c *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &AccessKeyMiddleware{TenantConfig: c}
	return m.Handle
}

func ProvideAuthContextGetter(ctx context.Context) ContextGetter {
	return NewContextGetterWithContext(ctx)
}

var DependencySet = wire.NewSet(ProvideAuthContextGetter)
