package authn

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type AuthContextResolverFactory interface {
	NewResolver(context.Context, config.TenantConfiguration) AuthContextResolver
}

type AuthContextResolver interface {
	Resolve(*http.Request, auth.ContextSetter) error
}
