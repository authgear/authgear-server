package authn

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type AuthContextResolverFactory interface {
	NewResolver(context.Context, config.TenantConfiguration) AuthContextResolver
}

type AuthContextResolver interface {
	Resolve(*http.Request) (handler.AuthContext, error)
}
