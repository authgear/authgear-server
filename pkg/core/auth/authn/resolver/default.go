package resolver

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authn"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContextResolverFactory struct {
	inject.ProviderGraph
}

func (f AuthContextResolverFactory) NewResolver(ctx context.Context, tenantConfig config.TenantConfiguration) authn.AuthContextResolver {
	r := &DefaultAuthContextResolver{}
	inject.DefaultInject(r, f.ProviderGraph, ctx, tenantConfig)
	return r
}

type DefaultAuthContextResolver struct {
	TokenStore    authtoken.Store `dependency:"TokenStore"`
	AuthInfoStore authinfo.Store  `dependency:"AuthInfoStore"`
}

func (r DefaultAuthContextResolver) Resolve(req *http.Request) (ctx handler.AuthContext, err error) {
	keyType := model.GetAccessKeyType(req)

	var resolver authn.AuthContextResolver
	if keyType == model.MasterAccessKey {
		resolver = masterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	} else {
		resolver = nonMasterkeyAuthContextResolver{
			TokenStore:    r.TokenStore,
			AuthInfoStore: r.AuthInfoStore,
		}
	}

	ctx, err = resolver.Resolve(req)
	ctx.AccessKeyType = keyType

	return
}
