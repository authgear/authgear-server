package authn

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	skyContext "github.com/skygeario/skygear-server/pkg/core/handler/context"
)

type AuthContextResolverFactory interface {
	NewResolver(context.Context, config.TenantConfiguration) AuthContextResolver
}

type AuthContextResolver interface {
	Resolve(*http.Request) (skyContext.AuthContext, error)
}
