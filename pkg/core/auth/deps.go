package auth

import (
	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideAccessKeyMiddleware(c *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &AccessKeyMiddleware{TenantConfig: c}
	return m.Handle
}
